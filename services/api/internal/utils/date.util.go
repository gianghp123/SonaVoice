package utils

import "time"

// NowUTC returns the current time in UTC.
// Use this for all timestamp fields to ensure consistent timezone handling.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// QuotaDate returns the current UTC date truncated to midnight.
// Use this for quota_date fields to ensure daily quota boundaries
// are calculated consistently regardless of server timezone.
func QuotaDate() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// IsStale returns true if t is older than the given duration from now.
// Use this for stale session checks instead of manual threshold math.
func IsStale(t time.Time, maxAge time.Duration) bool {
	return t.Before(NowUTC().Add(-maxAge))
}
