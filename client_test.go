package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeAPI returns a test server that mimics the real API envelope format.
func fakeAPI(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data":    map[string]string{"status": "ok"},
		})
	})

	// List plugins
	mux.HandleFunc("/v1/plugins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(w, map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"id":          "test-plugin",
						"name":        "Test Plugin",
						"description": "A test plugin",
						"category":    "cloud",
						"tags":        []string{"test"},
						"official":    true,
					},
				},
				"pagination": map[string]interface{}{
					"page":        1,
					"per_page":    10,
					"total":       1,
					"total_pages": 1,
				},
			})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Get plugin
	mux.HandleFunc("/v1/plugins/test-plugin", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":          "test-plugin",
				"name":        "Test Plugin",
				"description": "A test plugin",
				"category":    "cloud",
				"official":    true,
			},
		})
	})

	// List versions
	mux.HandleFunc("/v1/plugins/test-plugin/versions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{
					"id":        "ver-1",
					"plugin_id": "test-plugin",
					"version":   "1.0.0",
					"visible":   true,
					"artifacts": map[string]interface{}{
						"darwin_arm64": map[string]interface{}{
							"checksum":     "abc123",
							"signature":    "sig123",
							"download_url": "test-plugin/1.0.0/test-plugin-darwin_arm64.tar.gz",
							"size":         1024,
						},
					},
				},
			},
			"pagination": map[string]interface{}{
				"page":        1,
				"per_page":    10,
				"total":       1,
				"total_pages": 1,
			},
		})
	})

	// Get version
	mux.HandleFunc("/v1/plugins/test-plugin/versions/1.0.0", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":        "ver-1",
				"plugin_id": "test-plugin",
				"version":   "1.0.0",
				"visible":   true,
				"artifacts": map[string]interface{}{
					"darwin_arm64": map[string]interface{}{
						"checksum":     "abc123",
						"signature":    "sig123",
						"download_url": "test-plugin/1.0.0/test-plugin-darwin_arm64.tar.gz",
						"size":         1024,
					},
				},
			},
		})
	})

	// Categories
	mux.HandleFunc("/v1/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"category": "cloud", "count": 5},
				{"category": "database", "count": 3},
			},
		})
	})

	// Reviews
	mux.HandleFunc("/v1/plugins/test-plugin/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Check auth
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				writeJSON(w, map[string]interface{}{
					"success": false,
					"message": "authentication required",
				})
				return
			}
			writeJSON(w, map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":        "rev-1",
					"plugin_id": "test-plugin",
					"rating":    5,
					"title":     "Great",
					"body":      "Awesome plugin",
				},
			})
			return
		}
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{
					"id":        "rev-1",
					"plugin_id": "test-plugin",
					"rating":    5,
					"title":     "Great",
				},
			},
			"pagination": map[string]interface{}{
				"page": 1, "per_page": 10, "total": 1, "total_pages": 1,
			},
		})
	})

	// Publisher
	mux.HandleFunc("/v1/publishers/omniview", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":       "pub-1",
				"name":     "Omniview",
				"slug":     "omniview",
				"verified": true,
			},
		})
	})

	mux.HandleFunc("/v1/publishers/omniview/plugins", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"id": "test-plugin", "name": "Test Plugin"},
			},
			"pagination": map[string]interface{}{
				"page": 1, "per_page": 10, "total": 1, "total_pages": 1,
			},
		})
	})

	// Downloads
	mux.HandleFunc("/v1/plugins/test-plugin/downloads", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			writeJSON(w, map[string]interface{}{"success": true})
			return
		}
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"plugin_id":   "test-plugin",
				"total_count": 42,
				"by_version":  []map[string]interface{}{{"version": "1.0.0", "count": 42}},
			},
		})
	})

	mux.HandleFunc("/v1/plugins/test-plugin/downloads/daily", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"date": "2026-02-22", "count": 10},
				{"date": "2026-02-23", "count": 12},
			},
		})
	})

	// Submissions
	mux.HandleFunc("/v1/publishers/omniview/submissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(w, map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"id":           "sub-1",
						"publisher_id": "pub-1",
						"plugin_id":    "test-plugin",
						"version":      "2.0.0",
						"status":       "pending",
					},
				},
				"pagination": map[string]interface{}{
					"page": 1, "per_page": 10, "total": 1, "total_pages": 1,
				},
			})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v1/submissions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":           "sub-1",
				"publisher_id": "pub-1",
				"plugin_id":    "test-plugin",
				"version":      "2.0.0",
				"status":       "pending",
			},
		})
	})

	mux.HandleFunc("/v1/submissions/sub-1/withdraw", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			writeJSON(w, map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":           "sub-1",
					"publisher_id": "pub-1",
					"plugin_id":    "test-plugin",
					"version":      "2.0.0",
					"status":       "withdrawn",
				},
			})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// 404 catch-all
	mux.HandleFunc("/v1/plugins/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]interface{}{
			"success": false,
			"message": "plugin not found",
		})
	})

	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ─── Tests ──────────────────────────────────────────

func TestClient_Health(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	hs, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if hs.Status != "ok" {
		t.Fatalf("expected status ok, got %s", hs.Status)
	}
}

func TestClient_ListPlugins(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListPlugins(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPlugins() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Items))
	}
	if result.Items[0].ID != "test-plugin" {
		t.Fatalf("expected test-plugin, got %s", result.Items[0].ID)
	}
	if !result.Items[0].Official {
		t.Fatal("expected Official=true")
	}
	if result.Pagination == nil {
		t.Fatal("expected pagination")
	}
	if result.Pagination.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Pagination.Total)
	}
}

func TestClient_ListPlugins_withOptions(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListPlugins(context.Background(), &ListOptions{
		Page:     1,
		PerPage:  5,
		Category: "cloud",
		Search:   "test",
	})
	if err != nil {
		t.Fatalf("ListPlugins() error: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected at least 1 plugin")
	}
}

func TestClient_GetPlugin(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	p, err := c.GetPlugin(context.Background(), "test-plugin")
	if err != nil {
		t.Fatalf("GetPlugin() error: %v", err)
	}
	if p.ID != "test-plugin" {
		t.Fatalf("expected test-plugin, got %s", p.ID)
	}
	if p.Category != "cloud" {
		t.Fatalf("expected cloud category, got %s", p.Category)
	}
}

func TestClient_GetPlugin_notFound(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.GetPlugin(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound, got: %v", err)
	}
}

func TestClient_ListVersions(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListVersions(context.Background(), "test-plugin", nil)
	if err != nil {
		t.Fatalf("ListVersions() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 version, got %d", len(result.Items))
	}
	v := result.Items[0]
	if v.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", v.Version)
	}
	if len(v.Artifacts) == 0 {
		t.Fatal("expected artifacts map to be populated")
	}
	art, ok := v.Artifacts["darwin_arm64"]
	if !ok {
		t.Fatal("expected darwin_arm64 artifact")
	}
	if art.Checksum != "abc123" {
		t.Fatalf("expected checksum abc123, got %s", art.Checksum)
	}
	if art.Size != 1024 {
		t.Fatalf("expected size 1024, got %d", art.Size)
	}
}

func TestClient_GetVersion(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	v, err := c.GetVersion(context.Background(), "test-plugin", "1.0.0")
	if err != nil {
		t.Fatalf("GetVersion() error: %v", err)
	}
	if v.Version != "1.0.0" {
		t.Fatalf("expected 1.0.0, got %s", v.Version)
	}
	if v.Artifacts == nil {
		t.Fatal("expected artifacts")
	}
	if _, ok := v.Artifacts["darwin_arm64"]; !ok {
		t.Fatal("expected darwin_arm64 in artifacts")
	}
}

func TestClient_ListCategories(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	cats, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("ListCategories() error: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if cats[0].Category != "cloud" {
		t.Fatalf("expected cloud, got %s", cats[0].Category)
	}
}

func TestClient_ListReviews(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListReviews(context.Background(), "test-plugin", nil)
	if err != nil {
		t.Fatalf("ListReviews() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 review, got %d", len(result.Items))
	}
	if result.Items[0].Rating != 5 {
		t.Fatalf("expected rating 5, got %d", result.Items[0].Rating)
	}
}

func TestClient_CreateReview_authenticated(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithToken("test-token"))
	r, err := c.CreateReview(context.Background(), "test-plugin", &CreateReviewInput{
		Rating: 5,
		Title:  "Great",
		Body:   "Awesome plugin",
	})
	if err != nil {
		t.Fatalf("CreateReview() error: %v", err)
	}
	if r.Rating != 5 {
		t.Fatalf("expected rating 5, got %d", r.Rating)
	}
}

func TestClient_CreateReview_unauthenticated(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL)) // no token
	_, err := c.CreateReview(context.Background(), "test-plugin", &CreateReviewInput{
		Rating: 5,
	})
	if err == nil {
		t.Fatal("expected error for unauthenticated review")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestClient_GetPublisher(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	p, err := c.GetPublisher(context.Background(), "omniview")
	if err != nil {
		t.Fatalf("GetPublisher() error: %v", err)
	}
	if p.Slug != "omniview" {
		t.Fatalf("expected omniview slug, got %s", p.Slug)
	}
	if !p.Verified {
		t.Fatal("expected publisher to be verified")
	}
}

func TestClient_ListPublisherPlugins(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListPublisherPlugins(context.Background(), "omniview", nil)
	if err != nil {
		t.Fatalf("ListPublisherPlugins() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Items))
	}
}

func TestClient_GetDownloadStats(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	stats, err := c.GetDownloadStats(context.Background(), "test-plugin")
	if err != nil {
		t.Fatalf("GetDownloadStats() error: %v", err)
	}
	if stats.TotalCount != 42 {
		t.Fatalf("expected 42 total, got %d", stats.TotalCount)
	}
	if len(stats.ByVersion) != 1 {
		t.Fatalf("expected 1 version breakdown, got %d", len(stats.ByVersion))
	}
}

func TestClient_GetDailyDownloads(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	dd, err := c.GetDailyDownloads(context.Background(), "test-plugin", 7)
	if err != nil {
		t.Fatalf("GetDailyDownloads() error: %v", err)
	}
	if len(dd) != 2 {
		t.Fatalf("expected 2 days, got %d", len(dd))
	}
}

func TestClient_RecordDownload(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	err := c.RecordDownload(context.Background(), "test-plugin", "1.0.0", "darwin_arm64")
	if err != nil {
		t.Fatalf("RecordDownload() error: %v", err)
	}
}

func TestClient_ListSubmissions(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListSubmissions(context.Background(), "omniview", nil)
	if err != nil {
		t.Fatalf("ListSubmissions() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(result.Items))
	}
	if result.Items[0].ID != "sub-1" {
		t.Fatalf("expected sub-1, got %s", result.Items[0].ID)
	}
	if result.Items[0].Status != "pending" {
		t.Fatalf("expected pending status, got %s", result.Items[0].Status)
	}
}

func TestClient_ListSubmissions_withOptions(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	result, err := c.ListSubmissions(context.Background(), "omniview", &ListOptions{
		Status:   "pending",
		PluginID: "test-plugin",
		PerPage:  5,
	})
	if err != nil {
		t.Fatalf("ListSubmissions() error: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected at least 1 submission")
	}
}

func TestClient_GetSubmission(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	sub, err := c.GetSubmission(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("GetSubmission() error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Fatalf("expected sub-1, got %s", sub.ID)
	}
	if sub.PluginID != "test-plugin" {
		t.Fatalf("expected test-plugin, got %s", sub.PluginID)
	}
}

func TestClient_WithdrawSubmission(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	sub, err := c.WithdrawSubmission(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("WithdrawSubmission() error: %v", err)
	}
	if sub.Status != "withdrawn" {
		t.Fatalf("expected withdrawn status, got %s", sub.Status)
	}
}

func TestClient_WithHTTPClient(t *testing.T) {
	srv := fakeAPI(t)
	defer srv.Close()

	custom := &http.Client{}
	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(custom))
	_, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() with custom client error: %v", err)
	}
}

func TestClient_ReleaseChannels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/plugins/test-plugin/versions":
			if got := r.URL.Query().Get("channel"); got != "feature-abc" {
				t.Errorf("channel = %q", got)
			}
			writeJSON(w, map[string]interface{}{"success": true, "data": []map[string]string{{"version": "1.0.1-feature-abc.1.1", "channel": "feature-abc"}}})
		case "/v1/plugins/test-plugin/channels":
			writeJSON(w, map[string]interface{}{"success": true, "data": []map[string]string{{"channel": "feature-abc", "branch": "feature/abc", "status": "active", "created_at": "2026-01-01T00:00:00Z"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL))
	versions, err := c.ListVersions(context.Background(), "test-plugin", &ListOptions{Channel: "feature-abc"})
	if err != nil || len(versions.Items) != 1 || versions.Items[0].Channel != "feature-abc" {
		t.Fatalf("versions = %#v, error = %v", versions, err)
	}
	channels, err := c.ListChannels(context.Background(), "test-plugin")
	if err != nil || len(channels) != 1 || channels[0].Channel != "feature-abc" || channels[0].Branch != "feature/abc" || channels[0].CreatedAt.IsZero() {
		t.Fatalf("channels = %#v, error = %v", channels, err)
	}
}

func TestDownloadRanges(t *testing.T) {
	data := bytes.Repeat([]byte("a"), (16<<20)+3)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.ServeContent(w, r, "artifact", time.Time{}, bytes.NewReader(data))
	}))
	defer server.Close()
	var result bytes.Buffer
	err := NewClient().downloadRanges(context.Background(), server.URL, int64(len(data)), &result)
	if err != nil || !bytes.Equal(result.Bytes(), data) || requests != 2 {
		t.Fatalf("download: err=%v, bytes=%d, requests=%d", err, result.Len(), requests)
	}
}

func TestDownloadRangesRejectsInvalidResponses(t *testing.T) {
	for _, tc := range []struct {
		name, contentRange, body string
		status                   int
		size                     int64
	}{
		{"ignored range", "", "abc", http.StatusOK, 3},
		{"wrong start", "bytes 1-2/3", "bc", http.StatusPartialContent, 3},
		{"wrong size", "bytes 0-2/3", "abc", http.StatusPartialContent, 4},
		{"truncated body", "bytes 0-2/3", "ab", http.StatusPartialContent, 3},
		{"oversized body", "bytes 0-2/3", "abcd", http.StatusPartialContent, 3},
		{"trailing content range", "bytes 0-2/3 junk", "abc", http.StatusPartialContent, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.Header.Get("Range"), "bytes=0-") {
					t.Errorf("missing range: %q", r.Header.Get("Range"))
				}
				w.Header().Set("Content-Range", tc.contentRange)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			if err := NewClient().downloadRanges(context.Background(), server.URL, tc.size, &bytes.Buffer{}); err == nil {
				t.Fatal("expected invalid response to fail")
			}
		})
	}
}

func TestMultipartSubmissionMethods(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("X-API-Key") != "secret" {
			t.Errorf("unexpected method or authentication")
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["architecture"] != "linux_amd64" {
			t.Errorf("unexpected body: %v", body)
		}
		paths = append(paths, r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/init") {
			if body["part_count"] != float64(2) {
				t.Errorf("unexpected part count: %v", body)
			}
			writeJSON(w, map[string]interface{}{"success": true, "data": map[string]interface{}{"upload_id": "upload", "part_urls": []string{"one", "two"}}})
		} else {
			if body["upload_id"] != "upload" {
				t.Errorf("unexpected upload ID: %v", body)
			}
			if strings.HasSuffix(r.URL.Path, "/complete") {
				parts, ok := body["parts"].([]interface{})
				if !ok || len(parts) != 1 || parts[0].(map[string]interface{})["etag"] != `"etag"` {
					t.Errorf("unexpected parts: %v", body["parts"])
				}
			}
			writeJSON(w, map[string]interface{}{"success": true})
		}
	}))
	defer server.Close()
	client := NewClient(WithBaseURL(server.URL), WithAPIKey("secret"))
	ctx := context.Background()
	upload, err := client.InitMultipartUpload(ctx, "submission", "linux_amd64", 2)
	if err != nil || upload.UploadID != "upload" || len(upload.PartURLs) != 2 {
		t.Fatalf("init: %v, %v", upload, err)
	}
	if err := client.CompleteMultipartUpload(ctx, "submission", "linux_amd64", upload.UploadID, []MultipartPart{{PartNumber: 1, ETag: `"etag"`}}); err != nil {
		t.Fatal(err)
	}
	if err := client.AbortMultipartUpload(ctx, "submission", "linux_amd64", upload.UploadID); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || paths[0] != "/v1/submissions/submission/multipart/init" || paths[1] != "/v1/submissions/submission/multipart/complete" || paths[2] != "/v1/submissions/submission/multipart/abort" {
		t.Fatalf("unexpected paths: %v", paths)
	}
}
