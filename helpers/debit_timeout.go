package helpers

import "time"

func DebitTimeout(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}
