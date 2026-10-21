package mangoplus_test

import (
	"errors"
	"testing"

	"github.com/raf555/mangoplus"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()
	err := &mangoplus.APIError{StatusCode: 503}
	if got, want := err.Error(), "mangoplus: API returned non-200 status code: 503"; got != want {
		t.Errorf("APIError.Error() returned %q, want %q", got, want)
	}
}

func TestProtoError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  *mangoplus.ProtoError
		want string
	}{
		{
			name: "action, subject and body",
			err: &mangoplus.ProtoError{
				Action:       mangoplus.ErrorActionDefault,
				EnglishPopup: mangoplus.PopupOsDefault{Subject: "Not Found", Body: "This title does not exist."},
			},
			want: "mangoplus: proto error: DEFAULT: Not Found: This title does not exist.",
		},
		{
			name: "action only",
			err:  &mangoplus.ProtoError{Action: mangoplus.ErrorActionMaintenance},
			want: "mangoplus: proto error: MAINTENANCE",
		},
		{
			name: "action and body",
			err: &mangoplus.ProtoError{
				Action:       mangoplus.ErrorActionMaintenance,
				EnglishPopup: mangoplus.PopupOsDefault{Body: "Under maintenance."},
			},
			want: "mangoplus: proto error: MAINTENANCE: Under maintenance.",
		},
		{
			name: "subject and body without action",
			err: &mangoplus.ProtoError{
				EnglishPopup: mangoplus.PopupOsDefault{Subject: "Not Found", Body: "This title does not exist."},
			},
			want: "mangoplus: proto error: Not Found: This title does not exist.",
		},
		{
			name: "empty",
			err:  &mangoplus.ProtoError{},
			want: "mangoplus: proto error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ProtoError.Error() returned %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProtoError_Is(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		action  mangoplus.ErrorAction
		subject string
		target  error
		want    bool
	}{
		{"not found matches", "", "Not Found", mangoplus.ErrNotFound, true},
		{"not found mismatches invalid parameter", "", "Not Found", mangoplus.ErrInvalidParameter, false},
		{"invalid parameter matches", "", "Invalid Parameter", mangoplus.ErrInvalidParameter, true},
		{"new version available matches", "", "NEW VERSION AVAILABLE", mangoplus.ErrNewVersionAvailable, true},
		{"invalid user matches", "", "Invalid user", mangoplus.ErrInvalidUserAccess, true},
		{"unknown subject returns false", "", "Something Else", mangoplus.ErrNotFound, false},
		{"unrelated target returns false", "", "Not Found", errors.New("other"), false},
		{"maintenance action matches", mangoplus.ErrorActionMaintenance, "", mangoplus.ErrMaintenance, true},
		{"default action mismatches maintenance", mangoplus.ErrorActionDefault, "Not Found", mangoplus.ErrMaintenance, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &mangoplus.ProtoError{Action: tt.action, EnglishPopup: mangoplus.PopupOsDefault{Subject: tt.subject}}
			if got := errors.Is(err, tt.target); got != tt.want {
				t.Errorf("errors.Is(%+v, %v) = %t, want %t", err, tt.target, got, tt.want)
			}
		})
	}
}
