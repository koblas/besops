package heartbeat_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/koblas/besops/internal/database"
	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) (heartbeat.Repository, *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	dbURL := "sqlite://" + filepath.Join(dir, "test.db")

	db, err := database.Open(dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, database.Migrate(db, dbURL))

	return heartbeat.NewRepository(db), db
}

func createTestMonitor(t *testing.T, db *sql.DB) string {
	t.Helper()
	userID := uuid.New().String()
	username := "user-" + userID[:8]
	_, err := db.Exec(`INSERT INTO "user" (id, username, password) VALUES (?, ?, 'hashed')`, userID, username)
	require.NoError(t, err)

	id := uuid.New().String()
	_, err = db.Exec(`INSERT INTO monitor (id, name, type, active, user_id) VALUES (?, 'Test Monitor', 'http', 1, ?)`, id, userID)
	require.NoError(t, err)
	return id
}

func insertHeartbeat(t *testing.T, repo heartbeat.Repository, monitorID string, status int, important bool, at time.Time) {
	t.Helper()
	hb := &heartbeat.Heartbeat{
		MonitorID: monitorID,
		Status:    status,
		Time:      heartbeat.RFC3339Time(at),
		Important: important,
		Msg:       "test",
	}
	require.NoError(t, repo.Insert(context.Background(), hb))
}

func TestGetAllImportant_Empty(t *testing.T) {
	repo, _ := setupRepo(t)

	hbs, total, err := repo.GetAllImportant(t.Context(), 25)
	require.NoError(t, err)
	require.Empty(t, hbs)
	require.Equal(t, int64(0), total)
}

func TestGetAllImportant_ReturnsOnlyImportantHeartbeats(t *testing.T) {
	repo, db := setupRepo(t)
	monID := createTestMonitor(t, db)

	now := time.Now()
	insertHeartbeat(t, repo, monID, 1, false, now.Add(-3*time.Minute))
	insertHeartbeat(t, repo, monID, 0, true, now.Add(-2*time.Minute))
	insertHeartbeat(t, repo, monID, 1, true, now.Add(-1*time.Minute))

	hbs, total, err := repo.GetAllImportant(t.Context(), 25)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, hbs, 2)
}

func TestGetAllImportant_AcrossMultipleMonitors(t *testing.T) {
	repo, db := setupRepo(t)
	mon1 := createTestMonitor(t, db)
	mon2 := createTestMonitor(t, db)

	now := time.Now()
	insertHeartbeat(t, repo, mon1, 0, true, now.Add(-3*time.Minute))
	insertHeartbeat(t, repo, mon2, 0, true, now.Add(-2*time.Minute))
	insertHeartbeat(t, repo, mon1, 1, true, now.Add(-1*time.Minute))

	hbs, total, err := repo.GetAllImportant(t.Context(), 25)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, hbs, 3)
	// Newest first
	require.Equal(t, mon1, hbs[0].MonitorID)
	require.Equal(t, mon2, hbs[1].MonitorID)
}

func TestGetAllImportant_RespectsLimit(t *testing.T) {
	repo, db := setupRepo(t)
	monID := createTestMonitor(t, db)

	now := time.Now()
	for i := 0; i < 5; i++ {
		insertHeartbeat(t, repo, monID, i%2, true, now.Add(-time.Duration(5-i)*time.Minute))
	}

	hbs, total, err := repo.GetAllImportant(t.Context(), 3)
	require.NoError(t, err)
	require.Equal(t, int64(5), total)
	require.Len(t, hbs, 3)
}

func TestGetAllImportant_OrderedNewestFirst(t *testing.T) {
	repo, db := setupRepo(t)
	monID := createTestMonitor(t, db)

	now := time.Now()
	insertHeartbeat(t, repo, monID, 0, true, now.Add(-3*time.Minute))
	insertHeartbeat(t, repo, monID, 1, true, now.Add(-1*time.Minute))

	hbs, _, err := repo.GetAllImportant(t.Context(), 25)
	require.NoError(t, err)
	require.Len(t, hbs, 2)

	first := time.Time(hbs[0].Time)
	second := time.Time(hbs[1].Time)
	require.True(t, first.After(second), "expected newest first, got %v before %v", first, second)
}
