package queue

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
)

var AsynqClient *asynq.Client
var AsynqServer *asynq.Server
var Inspector *asynq.Inspector

func InitAsynq() {

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0,
	}

	AsynqClient = asynq.NewClient(redisOpt)

	Inspector = asynq.NewInspector(redisOpt)

	AsynqServer = asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
}

func StartWorker() {

	mux := asynq.NewServeMux()

	// register handlers
	mux.HandleFunc(TypeOrderQueueCash, HandleOrderQueueCashTask)
	mux.HandleFunc(TypeOrderOfferExpired, HandleOrderOfferExpiredTask)
	mux.HandleFunc(TypeOrderSelectedExpired, HandleOrderSelectedExpiredTask)
	mux.HandleFunc(TypeOrderOnProgressToFinish, HandleOrderOnProgressToFinishTask)
	mux.HandleFunc(TypeOrderEwalletNotifyExpired, HandleOrderEwalletNotifyExpiredTask)
	mux.HandleFunc(TypeOrderComingSoonRun, HandleOrderComingSoonRunTask)
	mux.HandleFunc(TypeOrderComingSoonWarning, HandleOrderComingSoonWarningTask)

	log.Println("Asynq worker started")

	if err := AsynqServer.Run(mux); err != nil {
		log.Fatalf("could not run task server: %v", err)
	}
}

func StopWorker() {
	if AsynqServer != nil {
		AsynqServer.Shutdown()
	}
	if AsynqClient != nil {
		AsynqClient.Close()
	}
	if Inspector != nil {
		Inspector.Close()
	}
}
