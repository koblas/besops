package jobs

import (
	"context"
	"fmt"

	"github.com/koblas/besops/internal/uptime"
)

// AggregateMinutelyJob rolls up heartbeats into per-minute stats.
type AggregateMinutelyJob struct {
	calc *uptime.Calculator
}

func NewAggregateMinutelyJob(calc *uptime.Calculator) *AggregateMinutelyJob {
	return &AggregateMinutelyJob{calc: calc}
}

func (j *AggregateMinutelyJob) Name() string     { return "aggregate_minutely" }
func (j *AggregateMinutelyJob) Schedule() string { return "0 * * * * *" } // every minute at :00

func (j *AggregateMinutelyJob) Run(ctx context.Context) error {
	if err := j.calc.AggregateMinutely(ctx); err != nil {
		return fmt.Errorf("minutely aggregation: %w", err)
	}
	return nil
}

// AggregateHourlyJob rolls up heartbeats into per-hour stats.
type AggregateHourlyJob struct {
	calc *uptime.Calculator
}

func NewAggregateHourlyJob(calc *uptime.Calculator) *AggregateHourlyJob {
	return &AggregateHourlyJob{calc: calc}
}

func (j *AggregateHourlyJob) Name() string     { return "aggregate_hourly" }
func (j *AggregateHourlyJob) Schedule() string { return "0 0 * * * *" } // every hour at :00:00

func (j *AggregateHourlyJob) Run(ctx context.Context) error {
	if err := j.calc.AggregateHourly(ctx); err != nil {
		return fmt.Errorf("hourly aggregation: %w", err)
	}
	return nil
}

// AggregateDailyJob rolls up heartbeats into per-day stats.
type AggregateDailyJob struct {
	calc *uptime.Calculator
}

func NewAggregateDailyJob(calc *uptime.Calculator) *AggregateDailyJob {
	return &AggregateDailyJob{calc: calc}
}

func (j *AggregateDailyJob) Name() string     { return "aggregate_daily" }
func (j *AggregateDailyJob) Schedule() string { return "0 0 0 * * *" } // daily at midnight

func (j *AggregateDailyJob) Run(ctx context.Context) error {
	if err := j.calc.AggregateDaily(ctx); err != nil {
		return fmt.Errorf("daily aggregation: %w", err)
	}
	return nil
}
