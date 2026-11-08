package mangoplus_test

import (
	"testing"

	"github.com/raf555/mangoplus"
)

func TestMaskSecretQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"keeps last 4", "app_ver=1&secret=0123456789abcdef&os=android", "app_ver=1&secret=****cdef&os=android"},
		{"short secret fully masked", "secret=s3cr3t", "secret=****"},
		{"empty secret", "secret=&os=android", "secret=****&os=android"},
		{"no secret", "os=android&os_ver=37", "os=android&os_ver=37"},
		{"similar key untouched", "device_secret=0123456789abcdef", "device_secret=0123456789abcdef"},
		{"empty query", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mangoplus.MaskSecretQuery(tt.in); got != tt.want {
				t.Errorf("MaskSecretQuery(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
