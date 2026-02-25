package registry

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "plugin not found"}
	got := err.Error()
	if got != "registry API error 404: plugin not found" {
		t.Fatalf("unexpected error string: %s", got)
	}
}

func TestIsNotFound_true(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "not found"}
	if !IsNotFound(err) {
		t.Fatal("expected IsNotFound to return true for 404")
	}
}

func TestIsNotFound_false(t *testing.T) {
	err := &APIError{StatusCode: 500, Message: "server error"}
	if IsNotFound(err) {
		t.Fatal("expected IsNotFound to return false for 500")
	}
}

func TestIsNotFound_nonAPIError(t *testing.T) {
	err := errors.New("something else")
	if IsNotFound(err) {
		t.Fatal("expected IsNotFound to return false for non-API error")
	}
}

func TestIsNotFound_wrappedAPIError(t *testing.T) {
	inner := &APIError{StatusCode: 404, Message: "wrapped"}
	wrapped := errors.Join(errors.New("context"), inner)
	if !IsNotFound(wrapped) {
		t.Fatal("expected IsNotFound to detect wrapped 404")
	}
}
