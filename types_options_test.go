package registry

import (
	"testing"
)

func TestListOptions_buildQuery_nil(t *testing.T) {
	var opts *ListOptions
	if got := opts.buildQuery(); got != "" {
		t.Fatalf("nil opts should produce empty query, got %q", got)
	}
}

func TestListOptions_buildQuery_empty(t *testing.T) {
	opts := &ListOptions{}
	if got := opts.buildQuery(); got != "" {
		t.Fatalf("empty opts should produce empty query, got %q", got)
	}
}

func TestListOptions_buildQuery_allFields(t *testing.T) {
	opts := &ListOptions{
		Page:           2,
		PerPage:        25,
		OrderField:     "created_at",
		OrderDirection: "desc",
		Search:         "kubernetes",
		Category:       "cloud",
		Featured:       true,
	}
	q := opts.buildQuery()
	if q == "" {
		t.Fatal("expected non-empty query string")
	}
	if q[0] != '?' {
		t.Fatalf("query should start with ?, got %q", q)
	}

	// Check that all params appear (order doesn't matter in URL encoding)
	checks := []string{
		"page=2",
		"per_page=25",
		"order_field=created_at",
		"order_direction=desc",
		"search=kubernetes",
		"category=cloud",
		"featured=true",
	}
	for _, c := range checks {
		if !contains(q, c) {
			t.Errorf("query %q missing %q", q, c)
		}
	}
}

func TestListOptions_buildQuery_partial(t *testing.T) {
	opts := &ListOptions{Page: 1, Search: "hello world"}
	q := opts.buildQuery()
	if !contains(q, "page=1") {
		t.Error("missing page param")
	}
	if !contains(q, "search=hello+world") && !contains(q, "search=hello%20world") {
		t.Error("missing or incorrectly encoded search param")
	}
	if contains(q, "per_page") {
		t.Error("per_page=0 should not appear")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
