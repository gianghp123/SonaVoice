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
