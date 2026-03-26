package queue

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
)

// NewTaskServer buat server Asynq, bisa dipakai di banyak scheduler
func NewTaskServer() *asynq.Server {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisHost + ":" + redisPort,
			Password: redisPassword,
			DB:       0,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	return srv
}

// RunTaskServer menjalankan server dengan mux handler
func RunTaskServer(mux *asynq.ServeMux) {
	srv := NewTaskServer()
	log.Println("Task server started")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run task server: %v", err)
	}
}
