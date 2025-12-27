package asynqinfra

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/user-service/internal/tasks"
)

type Service struct {
	server    *asynq.Server
	client    *asynq.Client
	scheduler *asynq.Scheduler
}

func New(
	redisConn asynq.RedisClientOpt,
	concurrency int,
) *Service {
	server := asynq.NewServer(
		redisConn,
		asynq.Config{Concurrency: concurrency},
	)

	client := asynq.NewClient(redisConn)

	scheduler := asynq.NewScheduler(redisConn, nil)

	return &Service{
		server:    server,
		client:    client,
		scheduler: scheduler,
	}
}

func (s *Service) Client() *asynq.Client {
	return s.client
}

func (s *Service) Start(mux *asynq.ServeMux) {
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

func (s *Service) Stop() {
	s.scheduler.Shutdown()
	log.Print("scheduler has been shutdown")

	s.server.Shutdown()
	log.Print("server has been shutdown")

	s.client.Close()
	log.Print("client has been closed")
}

func (s *Service) RegisterSchedules(cronSpec string) {
	if _, err := s.scheduler.Register(
		cronSpec,
		tasks.NewPasswordExpiryCheckTask(),
	); err != nil {
		log.Fatal(err)
	}
}
