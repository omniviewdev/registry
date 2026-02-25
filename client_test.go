package registry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
