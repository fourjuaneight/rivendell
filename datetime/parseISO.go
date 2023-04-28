package datetime

import (
	"time"
)

// Parse the given string in ISO 8601 format and return an instance of Date.
func ParseISO(isoString string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.999999-07:00"
	parsedDate, err := time.Parse(layout, isoString)

	if err != nil {
		return time.Time{}, err
	}

	return parsedDate, nil
}
