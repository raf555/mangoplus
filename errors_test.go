package mangoplus_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/raf555/mangoplus"
)

func TestAPIError_Error(t *testing.T) {
	e := &mangoplus.APIError{StatusCode: 503}
	got := e.Error()
	const want = "mangoplus: API returned non-200 status code: 503"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_Fields(t *testing.T) {
	headers := http.Header{"Content-Type": []string{"application/x-protobuf"}}
	body := []byte("raw-body")

	e := &mangoplus.APIError{
		StatusCode: 500,
		Headers:    headers,
		RawBody:    body,
	}

	if e.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", e.StatusCode)
	}
	if got := e.Headers.Get("Content-Type"); got != "application/x-protobuf" {
		t.Errorf("Headers[Content-Type] = %q, want %q", got, "application/x-protobuf")
	}
	if string(e.RawBody) != "raw-body" {
		t.Errorf("RawBody = %q, want %q", e.RawBody, "raw-body")
	}
}

func TestProtoError_Error(t *testing.T) {
	e := &mangoplus.ProtoError{
		EnglishPopup: mangoplus.PopupOsDefault{
			Subject: "Not Found",
			Body:    "This title does not exist.",
		},
	}
	const want = "mangoplus: proto error: Not Found: This title does not exist."
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestProtoError_Is(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		target  error
		want    bool
	}{
		{"not found matches", "Not Found", mangoplus.ErrNotFound, true},
		{"not found mismatches invalid parameter", "Not Found", mangoplus.ErrInvalidParameter, false},
		{"invalid parameter matches", "Invalid Parameter", mangoplus.ErrInvalidParameter, true},
		{"new version available matches", "NEW VERSION AVAILABLE", mangoplus.ErrNewVersionAvailable, true},
		{"invalid user matches", "Invalid user", mangoplus.ErrInvalidUserAccess, true},
		{"unknown subject returns false", "Something Else", mangoplus.ErrNotFound, false},
		{"unrelated target returns false", "Not Found", errors.New("other"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &mangoplus.ProtoError{EnglishPopup: mangoplus.PopupOsDefault{Subject: tt.subject}}
			if got := errors.Is(e, tt.target); got != tt.want {
				t.Errorf("errors.Is(%+v, %v) = %t, want %t", e, tt.target, got, tt.want)
			}
		})
	}
}
