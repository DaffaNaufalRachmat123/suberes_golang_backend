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

func GetStartEndDateFromString(date string) (string, string) {
	return date + " 00:00:00", date + " 23:59:59"
}

func IsScheduleDateValid(scheduleDateTimeStr string, timezoneCode string) (bool, error) {
	loc, err := time.LoadLocation(timezoneCode)
	if err != nil {
		return false, err
	}

	now := time.Now().In(loc)
	now = now.AddDate(0, 0, 1) // Add one day

	layout := "2006-01-02 15:04"
	scheduleDateTime, err := time.ParseInLocation(layout, scheduleDateTimeStr, loc)
	if err != nil {
		return false, err
	}

	return scheduleDateTime.After(now), nil
}
