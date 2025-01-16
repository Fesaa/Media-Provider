package services

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type CronService interface {
	NewJob(gocron.JobDefinition, gocron.Task, ...gocron.JobOption) (gocron.Job, error)
	RemoveJob(uuid.UUID) error
	Update(uuid.UUID, gocron.JobDefinition, gocron.Task, ...gocron.JobOption) (gocron.Job, error)
}

type cronService struct {
	gocron.Scheduler
	log zerolog.Logger
}

func CronServiceProvider(log zerolog.Logger) (CronService, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	s.Start()

	return cronService{
		Scheduler: s,
		log:       log,
	}, nil
}
