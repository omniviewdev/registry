package registry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeDeviceAuthAPI(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/auth/device/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"device_code":      "abc123def456abc123def456abc123def456abc1",
				"user_code":        "ABCD-EFGH",
				"verification_uri": "http://localhost:4321/device",
				"expires_in":       900,
				"interval":         5,
			},
		})
	})

	mux.HandleFunc("/v1/auth/device/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			DeviceCode string `json:"device_code"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		switch body.DeviceCode {
		case "pending-code":
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "authorization_pending",
			})
		case "slow-code":
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error":             "slow_down",
				"error_description": "polling too fast",
			})
		case "expired-code":
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "expired_token",
			})
		case "denied-code":
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "access_denied",
			})
		default: // success
			writeJSON(w, map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"access_token": "jwt-token-here",
					"token_type":   "Bearer",
					"expires_in":   259200,
				},
			})
		}
	})

	return httptest.NewServer(mux)
}

func TestDeviceAuthError_Error(t *testing.T) {
	t.Run("code only", func(t *testing.T) {
		err := &DeviceAuthError{Code: "authorization_pending"}
		if err.Error() != "device auth error: authorization_pending" {
			t.Fatalf("unexpected: %s", err.Error())
		}
	})

	t.Run("code with description", func(t *testing.T) {
		err := &DeviceAuthError{Code: "slow_down", Description: "polling too fast"}
		expected := "device auth error: slow_down — polling too fast"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestClient_DeviceAuthorize_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]interface{}{
			"success": false,
			"message": "internal failure",
		})
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceAuthorize(context.Background())
	if err == nil {
		t.Fatal("expected error from failed authorize")
	}
}

func TestClient_DeviceToken_nonRFCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		// Return a 400 without RFC 8628 error fields
		_, _ = w.Write([]byte("bad request body"))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected error for non-RFC 400")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", apiErr.StatusCode)
	}
}

func TestClient_DeviceToken_envelopeNotSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": false,
			"message": "something went wrong",
		})
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected error for non-success envelope")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
}

func TestClient_DeviceAuthorize(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.DeviceAuthorize(context.Background())
	if err != nil {
		t.Fatalf("DeviceAuthorize() error: %v", err)
	}
	if resp.DeviceCode == "" {
		t.Fatal("expected non-empty device_code")
	}
	if resp.UserCode == "" {
		t.Fatal("expected non-empty user_code")
	}
	if resp.VerificationURI == "" {
		t.Fatal("expected non-empty verification_uri")
	}
	if resp.ExpiresIn != 900 {
		t.Fatalf("expected expires_in=900, got %d", resp.ExpiresIn)
	}
	if resp.Interval != 5 {
		t.Fatalf("expected interval=5, got %d", resp.Interval)
	}
}

func TestClient_DeviceToken_success(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.DeviceToken(context.Background(), "success-code")
	if err != nil {
		t.Fatalf("DeviceToken() error: %v", err)
	}
	if resp.AccessToken != "jwt-token-here" {
		t.Fatalf("expected jwt-token-here, got %s", resp.AccessToken)
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("expected Bearer, got %s", resp.TokenType)
	}
}

func TestClient_DeviceToken_pending(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "pending-code")
	if err == nil {
		t.Fatal("expected error for pending token")
	}
	var deviceErr *DeviceAuthError
	if !errors.As(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T: %v", err, err)
	}
	if deviceErr.Code != "authorization_pending" {
		t.Fatalf("expected authorization_pending, got %s", deviceErr.Code)
	}
}

func TestClient_DeviceToken_slowDown(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "slow-code")
	if err == nil {
		t.Fatal("expected error for slow_down")
	}
	var deviceErr *DeviceAuthError
	if !errors.As(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T: %v", err, err)
	}
	if deviceErr.Code != "slow_down" {
		t.Fatalf("expected slow_down, got %s", deviceErr.Code)
	}
}

func TestClient_DeviceToken_expired(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "expired-code")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	var deviceErr *DeviceAuthError
	if !errors.As(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T: %v", err, err)
	}
	if deviceErr.Code != "expired_token" {
		t.Fatalf("expected expired_token, got %s", deviceErr.Code)
	}
}

func TestClient_DeviceToken_denied(t *testing.T) {
	srv := fakeDeviceAuthAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.DeviceToken(context.Background(), "denied-code")
	if err == nil {
		t.Fatal("expected error for denied token")
	}
	var deviceErr *DeviceAuthError
	if !errors.As(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T: %v", err, err)
	}
	if deviceErr.Code != "access_denied" {
		t.Fatalf("expected access_denied, got %s", deviceErr.Code)
	}
}
