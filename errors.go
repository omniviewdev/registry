package registry

import (
	"errors"
	"fmt"
)

var (
	// ErrUnsignedArtifact is returned when an artifact has no signature.
	ErrUnsignedArtifact = errors.New("artifact is not signed")

	// ErrInvalidSignature is returned when an artifact signature fails verification.
	ErrInvalidSignature = errors.New("invalid artifact signature")

	// ErrNoPlatformArtifact is returned when no artifact exists for the current platform.
	ErrNoPlatformArtifact = errors.New("no artifact available for current platform")

	// ErrChecksumMismatch is returned when a downloaded file's checksum doesn't match.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrEmptyVersion is returned when a version string is empty.
	ErrEmptyVersion = errors.New("version string is empty")
)

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("registry API error %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 API error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}
