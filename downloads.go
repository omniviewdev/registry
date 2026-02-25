package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// GetDownloadURL returns the download URL for a specific plugin version and architecture.
func (c *Client) GetDownloadURL(ctx context.Context, pluginID, version, arch string) (string, error) {
	path := fmt.Sprintf("/v1/plugins/%s/download/%s/%s", pluginID, version, arch)
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Don't follow redirects — we want the Location header
	client := *c.httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
		return resp.Header.Get("Location"), nil
	}

	if resp.StatusCode >= 400 {
		return "", &APIError{StatusCode: resp.StatusCode, Message: "download not available"}
	}

	return "", fmt.Errorf("unexpected status %d from download endpoint", resp.StatusCode)
}

// RecordDownload records a download event for analytics.
func (c *Client) RecordDownload(ctx context.Context, pluginID, version, arch string) error {
	body := map[string]string{
		"plugin_id": pluginID,
		"version":   version,
		"arch":      arch,
	}
	return c.post(ctx, fmt.Sprintf("/v1/plugins/%s/downloads", pluginID), body, nil)
}

// GetDownloadStats returns aggregate download stats for a plugin.
func (c *Client) GetDownloadStats(ctx context.Context, pluginID string) (*DownloadStats, error) {
	var ds DownloadStats
	if err := c.get(ctx, fmt.Sprintf("/v1/plugins/%s/downloads", pluginID), &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}

// GetDailyDownloads returns daily download counts for a plugin.
func (c *Client) GetDailyDownloads(ctx context.Context, pluginID string, days int) ([]DailyDownloads, error) {
	path := fmt.Sprintf("/v1/plugins/%s/downloads/daily?days=%s", pluginID, strconv.Itoa(days))
	var dd []DailyDownloads
	if err := c.get(ctx, path, &dd); err != nil {
		return nil, err
	}
	return dd, nil
}

// DownloadPlugin downloads, verifies, and returns the temp file path for a plugin version.
// It auto-detects the current platform architecture.
func (c *Client) DownloadPlugin(ctx context.Context, pluginID, version string) (string, error) {
	// 1. Get version info with artifacts
	v, err := c.GetVersion(ctx, pluginID, version)
	if err != nil {
		return "", fmt.Errorf("getting version info: %w", err)
	}

	// 2. Determine current platform
	platform := CurrentPlatform()

	// 3. Look up artifact
	artifact, ok := v.Artifacts[platform]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrNoPlatformArtifact, platform)
	}

	// 4. Get download URL (follows redirect)
	downloadURL, err := c.GetDownloadURL(ctx, pluginID, version, platform)
	if err != nil {
		return "", fmt.Errorf("getting download URL: %w", err)
	}

	// 5. Download to temp file, computing SHA256 during download
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("omniview-plugin-%s-%s-*.tar.gz", pluginID, version))
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("creating download request: %w", err)
	}

	dlResp, err := c.httpClient.Do(dlReq)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("downloading artifact: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", &APIError{StatusCode: dlResp.StatusCode, Message: "download failed"}
	}

	hasher := sha256.New()
	w := io.MultiWriter(tmpFile, hasher)
	if _, err := io.Copy(w, dlResp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("writing artifact: %w", err)
	}
	tmpFile.Close()

	// 6. Verify checksum
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if checksum != artifact.Checksum {
		os.Remove(tmpPath)
		return "", fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, artifact.Checksum, checksum)
	}

	// 7. Verify signature
	if err := VerifyArtifactSignature(checksum, artifact.Signature); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("signature verification failed: %w", err)
	}

	return tmpPath, nil
}
