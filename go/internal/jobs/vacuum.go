package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
)

// VacuumJob runs SQLite VACUUM to reclaim disk space.
type VacuumJob struct {
	db *sql.DB
}

func NewVacuumJob(db *sql.DB) *VacuumJob {
	return &VacuumJob{db: db}
}

func (j *VacuumJob) Name() string     { return "vacuum" }
func (j *VacuumJob) Schedule() string { return "0 0 2 * * *" } // daily at 02:00

func (j *VacuumJob) Run(ctx context.Context) error {
	if _, err := j.db.ExecContext(ctx, "VACUUM"); err != nil {
		return fmt.Errorf("vacuuming database: %w", err)
	}
	runtime.GC()
	return nil
}
