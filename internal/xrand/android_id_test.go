package xrand_test

import (
	"encoding/hex"
	"testing"

	"github.com/raf555/mangoplus/internal/xrand"
)

func TestGenerateAndroidID(t *testing.T) {
	id, err := xrand.GenerateAndroidID()
	if err != nil {
		t.Fatalf("GenerateAndroidID() returned error: %v", err)
	}

	// 8 random bytes hex-encoded => 16 hex characters.
	const wantLen = 16
	if len(id) != wantLen {
		t.Errorf("len(id) = %d, want %d (id=%q)", len(id), wantLen, id)
	}

	if _, err := hex.DecodeString(id); err != nil {
		t.Errorf("id is not valid hex: %v (id=%q)", err, id)
	}
}

func TestGenerateAndroidID_Uniqueness(t *testing.T) {
	// Two independent calls should overwhelmingly return different values.
	// Collision probability for 64 random bits is negligible.
	a, err := xrand.GenerateAndroidID()
	if err != nil {
		t.Fatalf("GenerateAndroidID() first call error: %v", err)
	}

	b, err := xrand.GenerateAndroidID()
	if err != nil {
		t.Fatalf("GenerateAndroidID() second call error: %v", err)
	}

	if a == b {
		t.Errorf("GenerateAndroidID() returned identical values twice: %q", a)
	}
}
