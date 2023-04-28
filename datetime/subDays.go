package datetime

import (
	"time"
)

// Subtract the specified number of days from the given date.
func SubDays(date time.Time, days int) time.Time {
	return date.AddDate(0, 0, -days)
}
