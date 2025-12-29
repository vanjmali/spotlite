package asynqinfra

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/user-service/internal/tasks"
)

type AsynqService struct {
	server    *asynq.Server
	client    *asynq.Client
	scheduler *asynq.Scheduler
}

// New function initializes Asynq server, client and scheduler.
func New(
	redisConn asynq.RedisClientOpt,
	concurrency int,
) *AsynqService {
	server := asynq.NewServer(
		redisConn,
		asynq.Config{Concurrency: concurrency},
	)

	client := asynq.NewClient(redisConn)

	scheduler := asynq.NewScheduler(redisConn, nil)

	return &AsynqService{
		server:    server,
		client:    client,
		scheduler: scheduler,
	}
}

// Client function provides an interface to access the client.
func (s *AsynqService) Client() *asynq.Client {
	return s.client
}

// Start function starts the Asynq server and scheduler in two different goroutines.
func (s *AsynqService) Start(mux *asynq.ServeMux) {
	go func() {
		if err := s.server.Run(mux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := s.scheduler.Run(); err != nil {
			log.Fatal(err)
		}
	}()
}

// Stop function stops the scheduler, server and closes the client
// the order of execution is very important.
func (s *AsynqService) Stop() error {
	s.scheduler.Shutdown()
	log.Print("INFO: scheduler has been shutdown")

	s.server.Shutdown()
	log.Print("INFO: server has been shutdown")

	err := s.client.Close()
	if err != nil {
		log.Print("ERROR: An error has occurred while closing Asynq client")
		return err
	}

	log.Print("INFO: client has been closed")

	return nil
}

// RegisterSchedule initializes and registers a new scheduler.
func (s *AsynqService) RegisterSchedule(cronSpec string) {
	if _, err := s.scheduler.Register(
		cronSpec,
		tasks.NewPasswordExpiryCheckTask(),
	); err != nil {
		log.Fatal(err)
	}
}
