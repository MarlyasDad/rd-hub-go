package scheduler

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"time"
)

func NewScheduler() (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler: s,
	}, nil
}

type Scheduler struct {
	scheduler gocron.Scheduler
}

func (s *Scheduler) NewDurationJob(duration time.Duration, job any) (uuid.UUID, error) {
	j, err := s.scheduler.NewJob(
		gocron.DurationJob(duration),
		gocron.NewTask(job),
	)
	if err != nil {
		return uuid.UUID{}, err
	}

	return j.ID(), nil
}

func (s *Scheduler) NewCronJob(crontab string, job any) (uuid.UUID, error) {
	j, err := s.scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(job),
	)
	if err != nil {
		return uuid.UUID{}, err
	}

	return j.ID(), nil
}

func (s *Scheduler) RemoveJob(jobID uuid.UUID) error {
	return s.scheduler.RemoveJob(jobID)
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
}

func (s *Scheduler) Shutdown() error {
	return s.scheduler.Shutdown()
}
