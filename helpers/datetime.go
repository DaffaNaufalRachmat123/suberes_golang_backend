package helpers

import (
	"time"
)

func GetStartAndEndDate(timezoneCode string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezoneCode)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	now := time.Now().In(loc)
	startDateTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endDateTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)

	return startDateTime, endDateTime, nil
}
