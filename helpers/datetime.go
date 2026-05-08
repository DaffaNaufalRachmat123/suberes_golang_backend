package helpers

import (
	"errors"
	"fmt"
	"time"
)

// TimezoneLabel returns WIB/WITA/WIT label for the given IANA timezone code.
func TimezoneLabel(timezoneCode string) string {
	switch timezoneCode {
	case "Asia/Jakarta":
		return "WIB"
	case "Asia/Makassar":
		return "WITA"
	case "Asia/Jayapura":
		return "WIT"
	default:
		return "WIB"
	}
}

// FormatDateByTimezone formats t into "DD-MM-YYYY HH:MM:SS WIB/WITA/WIT" based on timezoneCode.
// Falls back to Asia/Jakarta if timezoneCode is invalid.
func FormatDateByTimezone(t time.Time, timezoneCode string) string {
	iana := timezoneCode
	if iana == "" {
		iana = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(iana)
	if err != nil {
		loc, _ = time.LoadLocation("Asia/Jakarta")
	}
	label := TimezoneLabel(iana)
	tLocal := t.In(loc)
	// Format bulan pakai nama Indonesia (Mei, dst)
	monthIndo := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	month := monthIndo[int(tLocal.Month())-1]
	return fmt.Sprintf("%02d %s %d %02d:%02d:%02d %s", tLocal.Day(), month, tLocal.Year(), tLocal.Hour(), tLocal.Minute(), tLocal.Second(), label)
}

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
