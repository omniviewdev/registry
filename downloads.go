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
// It auto-detects the current platform architecture and downloads in 16 MiB
// ranges. The artifact server must support HTTP range requests.
func (c *Client) DownloadPlugin(ctx context.Context, pluginID, version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("%w: plugin %q", ErrEmptyVersion, pluginID)
	}

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

	hasher := sha256.New()
	w := io.MultiWriter(tmpFile, hasher)
	if err := c.downloadRanges(ctx, downloadURL, artifact.Size, w); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("downloading artifact: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("closing artifact: %w", err)
	}

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

// downloadRanges bounds every response and validates that the server returned
// exactly the requested bytes from one consistently sized object.
func (c *Client) downloadRanges(ctx context.Context, url string, expectedSize int64, dst io.Writer) error {
	const chunkSize int64 = 16 << 20
	var offset, total int64
	for {
		end := offset + chunkSize - 1
		if total > 0 && end >= total {
			end = total - 1
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, end))
		req.Header.Set("Accept-Encoding", "identity")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		var start, last, size int64
		if resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			return &APIError{StatusCode: resp.StatusCode, Message: "range download failed"}
		}
		contentRange := resp.Header.Get("Content-Range")
		n, parseErr := fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &last, &size)
		if parseErr != nil || n != 3 || contentRange != fmt.Sprintf("bytes %d-%d/%d", start, last, size) ||
			start != offset || size <= 0 || last < start || last >= size ||
			last != min(end, size-1) || (total != 0 && size != total) ||
			(expectedSize > 0 && size != expectedSize) {
			resp.Body.Close()
			return fmt.Errorf("invalid Content-Range %q", contentRange)
		}
		total = size
		length := last - start + 1
		written, copyErr := io.Copy(dst, io.LimitReader(resp.Body, length+1))
		closeErr := resp.Body.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if written != length {
			return fmt.Errorf("range body has %d bytes, expected %d", written, length)
		}
		offset += length
		if offset == total {
			return nil
		}
	}
}
