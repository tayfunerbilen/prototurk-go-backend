package utils

import "time"

// Now returns the current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}

// ParseTime parses a time string and ensures it's in UTC
func ParseTime(t time.Time) time.Time {
	return t.UTC()
}
