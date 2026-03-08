package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DeviceAuthInitiateResponse is returned when starting the device flow.
type DeviceAuthInitiateResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceAuthTokenResponse is returned on successful token exchange.
type DeviceAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// DeviceAuthError represents an RFC 8628 error response.
type DeviceAuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

func (e *DeviceAuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("device auth error: %s — %s", e.Code, e.Description)
	}
	return fmt.Sprintf("device auth error: %s", e.Code)
}

// DeviceAuthorize initiates the device authorization flow.
func (c *Client) DeviceAuthorize(ctx context.Context) (*DeviceAuthInitiateResponse, error) {
	var resp DeviceAuthInitiateResponse
	if err := c.post(ctx, "/v1/auth/device/authorize", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeviceToken polls for a token using the device code.
// Returns (*DeviceAuthTokenResponse, nil) on success.
// Returns (nil, *DeviceAuthError) for RFC 8628 errors (authorization_pending, slow_down, etc.).
// Returns (nil, error) for unexpected errors.
func (c *Client) DeviceToken(ctx context.Context, deviceCode string) (*DeviceAuthTokenResponse, error) {
	body := map[string]string{
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := c.baseURL + "/v1/auth/device/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytesReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// RFC 8628 errors come as 400 with error/error_description fields
	if resp.StatusCode >= 400 {
		var deviceErr DeviceAuthError
		if json.Unmarshal(respBody, &deviceErr) == nil && deviceErr.Code != "" {
			return nil, &deviceErr
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	// Success — response is wrapped in our standard envelope
	var envelope apiResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !envelope.Success {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: envelope.Message}
	}

	var tokenResp DeviceAuthTokenResponse
	if envelope.Data != nil {
		if err := json.Unmarshal(envelope.Data, &tokenResp); err != nil {
			return nil, fmt.Errorf("decoding token response: %w", err)
		}
	}
	return &tokenResp, nil
}
