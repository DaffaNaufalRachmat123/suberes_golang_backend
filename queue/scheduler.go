package queue

import (
	"errors"
	"log"
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
		log.Printf("ScheduleOnceAt: waktu %v sudah lewat, task akan dijalankan segera", runAt)
		delay = 0
	}

	task := asynq.NewTask(taskType, payload)
	_, err := AsynqClient.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		log.Printf("failed to enqueue task %s: %v", taskType, err)
	}
	return err
}

// ScheduleOnceWithCallbackAt mirip node-schedule
func ScheduleOnceWithCallbackAt(runAt time.Time, callback func() error) {
	delay := time.Until(runAt)
	if delay < 0 {
		log.Printf("ScheduleOnceWithCallbackAt: waktu %v sudah lewat, menjalankan segera", runAt)
		delay = 0
	}

	go func() {
		time.Sleep(delay)
		if err := callback(); err != nil {
			log.Printf("task callback failed: %v", err)
		}
	}()
}
