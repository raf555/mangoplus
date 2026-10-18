package timex_test

import (
	"testing"
	"time"

	"github.com/raf555/mangoplus/internal/timex"
)

func TestUnix(t *testing.T) {
	t.Run("zero timestamp returns zero time", func(t *testing.T) {
		got := timex.Unix(0)
		if !got.IsZero() {
			t.Errorf("Unix(0) = %v, want zero time", got)
		}
	})

	t.Run("non-zero timestamp", func(t *testing.T) {
		const ts int64 = 1_700_000_000
		got := timex.Unix(ts)
		if got.Unix() != ts {
			t.Errorf("Unix(%d).Unix() = %d, want %d", ts, got.Unix(), ts)
		}
		if !got.Equal(time.Unix(ts, 0)) {
			t.Errorf("Unix(%d) = %v, want %v", ts, got, time.Unix(ts, 0))
		}
	})

	t.Run("negative timestamp", func(t *testing.T) {
		const ts int64 = -1
		got := timex.Unix(ts)
		if got.Unix() != ts {
			t.Errorf("Unix(%d).Unix() = %d, want %d", ts, got.Unix(), ts)
		}
	})
}
