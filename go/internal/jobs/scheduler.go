package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a background task that runs on a schedule.
type Job interface {
	Name() string
	Schedule() string
	Run(ctx context.Context) error
}

// Scheduler manages the lifecycle of background jobs using cron.
type Scheduler struct {
	cron *cron.Cron
	ctx  context.Context
	jobs []Job
}

// NewScheduler creates a scheduler that runs jobs with second-level precision.
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

// Register adds a job to the scheduler. Must be called before Start.
func (s *Scheduler) Register(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start begins executing all registered jobs on their schedules.
func (s *Scheduler) Start(ctx context.Context) error {
	s.ctx = ctx

	for _, j := range s.jobs {
		job := j
		_, err := s.cron.AddFunc(job.Schedule(), func() {
			if runErr := job.Run(s.ctx); runErr != nil {
				slog.ErrorContext(s.ctx, "job failed", slog.String("job", job.Name()), slog.Any("error", runErr))
			}
		})
		if err != nil {
			return fmt.Errorf("registering job %q: %w", job.Name(), err)
		}
		slog.InfoContext(ctx, "registered job", slog.String("job", job.Name()), slog.String("schedule", job.Schedule()))
	}

	s.cron.Start()
	return nil
}

// Stop gracefully stops the scheduler, waiting up to 5 seconds for running jobs to complete.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}
}
