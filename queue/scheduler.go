package queue

import (
	"errors"
	"time"

	"github.com/hibiken/asynq"
)

var ErrAsynqClientNotInitialized = errors.New("asynq client not initialized")

// ScheduleOnceAt jadwal task Asynq di waktu tertentu
func ScheduleOnceAt(taskType string, payload []byte, runAt time.Time) error {
	if AsynqClient == nil {
		return ErrAsynqClientNotInitialized
	}

	delay := time.Until(runAt)
	if delay < 0 {
		delay = 0
	}

	task := asynq.NewTask(taskType, payload)
	_, err := AsynqClient.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
	}
	return err
}

// ScheduleOnceWithCallbackAt mirip node-schedule
func ScheduleOnceWithCallbackAt(runAt time.Time, callback func() error) {
	delay := time.Until(runAt)
	if delay < 0 {
		delay = 0
	}

	go func() {
		time.Sleep(delay)
		if err := callback(); err != nil {
		}
	}()
}
