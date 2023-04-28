package datetime

import (
	"time"
)

// Is the first date after the second one?
func IsAfter(date1, date2 time.Time) bool {
	return date1.After(date2)
}
