package registry

import (
	"runtime"
	"testing"
)

func TestCurrentPlatform(t *testing.T) {
	got := CurrentPlatform()
	expected := runtime.GOOS + "_" + runtime.GOARCH
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestSupportedPlatforms(t *testing.T) {
	if len(SupportedPlatforms) == 0 {
		t.Fatal("SupportedPlatforms should not be empty")
	}

	seen := make(map[string]bool)
	for _, p := range SupportedPlatforms {
		if seen[p] {
			t.Fatalf("duplicate platform: %s", p)
		}
		seen[p] = true
	}

	// Current platform should be in the list (assuming test runs on a supported platform)
	current := CurrentPlatform()
	if !seen[current] {
		t.Logf("warning: current platform %s not in SupportedPlatforms (may be expected in CI)", current)
	}
}
