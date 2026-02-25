package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.omniview.dev"
	defaultTimeout = 30 * time.Second
)

// Client is the registry API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets the API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithToken sets the JWT bearer token for authenticated requests.
func WithToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a new registry API client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Health checks the API health endpoint.
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	var hs HealthStatus
	if err := c.get(ctx, "/v1/health", &hs); err != nil {
		return nil, err
	}
	return &hs, nil
}

// get performs a GET request and decodes the response data into dst.
func (c *Client) get(ctx context.Context, path string, dst interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, dst)
}

// post performs a POST request with a JSON body and decodes the response data into dst.
func (c *Client) post(ctx context.Context, path string, body interface{}, dst interface{}) error {
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		r = bytesReader(data)
	}
	return c.do(ctx, http.MethodPost, path, r, dst)
}

// do performs an HTTP request and decodes the API envelope response.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, dst interface{}) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var envelope apiResponse
		if json.Unmarshal(respBody, &envelope) == nil && envelope.Message != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: envelope.Message}
		}
		return &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if dst == nil {
		return nil
	}

	var envelope apiResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return fmt.Errorf("decoding response envelope: %w", err)
	}
	if !envelope.Success {
		return &APIError{StatusCode: resp.StatusCode, Message: envelope.Message}
	}
	if envelope.Data != nil {
		if err := json.Unmarshal(envelope.Data, dst); err != nil {
			return fmt.Errorf("decoding response data: %w", err)
		}
	}
	return nil
}

// getList performs a GET and decodes a paginated list response.
func (c *Client) getList(ctx context.Context, path string, dst interface{}) (*Pagination, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var envelope apiResponse
		if json.Unmarshal(respBody, &envelope) == nil && envelope.Message != "" {
			return nil, &APIError{StatusCode: resp.StatusCode, Message: envelope.Message}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	var envelope apiResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("decoding response envelope: %w", err)
	}
	if !envelope.Success {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: envelope.Message}
	}
	if envelope.Data != nil {
		if err := json.Unmarshal(envelope.Data, dst); err != nil {
			return nil, fmt.Errorf("decoding response data: %w", err)
		}
	}
	return envelope.Pagination, nil
}

// bytesReader wraps a byte slice as an io.Reader.
func bytesReader(b []byte) io.Reader {
	return &bytesReaderImpl{data: b}
}

type bytesReaderImpl struct {
	data []byte
	pos  int
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
