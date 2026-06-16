package timex

import "time"

// Unix returns [time.Time] provided ts.
// If ts is zero, Unix returns zero [time.Time].
func Unix(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}

	return time.Unix(ts, 0)
}
