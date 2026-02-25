package registry

import "runtime"

// SupportedPlatforms lists the platforms for which plugin artifacts are built.
var SupportedPlatforms = []string{
	"darwin_arm64",
	"darwin_amd64",
	"linux_amd64",
	"linux_arm64",
	"windows_amd64",
}

// CurrentPlatform returns the platform string for the current OS and architecture
// in the format used by artifact keys (e.g. "darwin_arm64").
func CurrentPlatform() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}
