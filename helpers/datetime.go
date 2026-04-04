package helpers

import (
	"errors"
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

	layouts := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	var scheduleDateTime time.Time
	var parseErr error
	for _, layout := range layouts {
		scheduleDateTime, parseErr = time.ParseInLocation(layout, scheduleDateTimeStr, loc)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		return false, parseErr
	}

	// Must not be in the past or equal to current time
	if !scheduleDateTime.After(now) {
		return false, nil
	}

	// Time of day must be between 06:00 and 23:59
	totalMinutes := scheduleDateTime.Hour()*60 + scheduleDateTime.Minute()
	if totalMinutes < 6*60 || totalMinutes > 23*60+59 {
		return false, errors.New("schedule time must be between 06:00 and 23:59")
	}

	return true, nil
}
