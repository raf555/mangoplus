package encodingx_test

import (
	"testing"

	"github.com/raf555/mangoplus/internal/encodingx"
)

func TestMD5Hex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string",
			in:   "",
			want: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name: "ascii",
			in:   "abc",
			want: "900150983cd24fb0d6963f7d28e17f72",
		},
		{
			name: "known vector",
			in:   "The quick brown fox jumps over the lazy dog",
			want: "9e107d9d372bb6826bd81d3542a419d6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodingx.MD5Hex(tt.in)
			if got != tt.want {
				t.Errorf("MD5Hex(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
