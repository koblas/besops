package system

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

const version = "2.0.0-beta"

var _ oas.SystemHandler = (*Handler)(nil)

type StatsClearer interface {
	ClearAll(ctx context.Context) error
}

type Handler struct {
	db    *sql.DB
	stats StatsClearer
}

func NewHandler(db *sql.DB, stats StatsClearer) *Handler {
	return &Handler{db: db, stats: stats}
}

func (h *Handler) HealthCheck(ctx context.Context) (oas.HealthCheckRes, error) {
	return &oas.HealthCheckOK{Status: oas.HealthCheckOKStatusOk}, nil
}

func (h *Handler) GetInfo(ctx context.Context) (oas.GetInfoRes, error) {
	tz := time.Now().Location().String()
	_, offset := time.Now().Zone()
	tzOffset := time.Duration(offset) * time.Second

	return &oas.GetInfoOK{
		Version:              oas.NewOptString(version),
		PrimaryBaseURL:       oas.OptURI{},
		ServerTimezone:       oas.NewOptString(tz),
		ServerTimezoneOffset: oas.NewOptString(tzOffset.String()),
	}, nil
}

func (h *Handler) GetDatabaseSize(ctx context.Context) (oas.GetDatabaseSizeRes, error) {
	var size int64
	row := h.db.QueryRowContext(ctx, `SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()`)
	if err := row.Scan(&size); err != nil {
		fi, statErr := os.Stat("data/besops.db")
		if statErr == nil {
			size = fi.Size()
		}
	}
	return &oas.GetDatabaseSizeOK{Size: size}, nil
}

func (h *Handler) ShrinkDatabase(ctx context.Context) (oas.ShrinkDatabaseRes, error) {
	if _, err := h.db.ExecContext(ctx, "VACUUM"); err != nil {
		return nil, fmt.Errorf("vacuuming database: %w", err)
	}
	runtime.GC()
	return &oas.MessageResponse{Message: "done"}, nil
}

func (h *Handler) ClearStatistics(ctx context.Context) (oas.ClearStatisticsRes, error) {
	if err := h.stats.ClearAll(ctx); err != nil {
		return nil, fmt.Errorf("clearing statistics: %w", err)
	}
	return &oas.ClearStatisticsNoContent{}, nil
}
