package datetime

import (
	"time"
)

// Subtract the specified number of hours from the given date.
func SubHours(date time.Time, hours int) time.Time {
	return date.Add(time.Duration(-hours) * time.Hour)
}
