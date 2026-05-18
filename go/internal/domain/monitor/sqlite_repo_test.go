package monitor_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/koblas/besops/internal/database"
	"github.com/koblas/besops/internal/domain/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) (monitor.Repository, *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	dbURL := "sqlite://" + filepath.Join(dir, "test.db")

	db, err := database.Open(dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, database.Migrate(db, dbURL))

	return monitor.NewRepository(db), db
}

// createUser inserts a user row and returns the user ID.
func createUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := uuid.New().String()
	username := "user-" + id[:8]
	_, err := db.Exec(`INSERT INTO "user" (id, username, password) VALUES (?, ?, 'hashed')`, id, username)
	require.NoError(t, err)
	return id
}

// createMonitor inserts a minimal monitor and returns its ID.
func createMonitor(t *testing.T, db *sql.DB, userID string, name string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO monitor (id, name, type, active, user_id) VALUES (?, ?, 'http', 1, ?)`,
		id, name, userID,
	)
	require.NoError(t, err)
	return id
}

// createTag inserts a tag and returns its ID.
func createTag(t *testing.T, db *sql.DB, name string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO tag (id, name, color, created_date) VALUES (?, ?, '#000000', datetime('now'))`,
		id, name,
	)
	require.NoError(t, err)
	return id
}

// linkMonitorTag inserts a row into monitor_tag.
func linkMonitorTag(t *testing.T, db *sql.DB, monitorID, tagID string) {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO monitor_tag (id, tag_id, monitor_id, value) VALUES (?, ?, ?, '')`,
		id, tagID, monitorID,
	)
	require.NoError(t, err)
}

// TestFindByTagIDs_ReturnsCorrectMonitors verifies that only monitors
// associated with the queried tag IDs are returned.
func TestFindByTagIDs_ReturnsCorrectMonitors(t *testing.T) {
	// Given: 3 monitors where monitor-A and monitor-B are tagged, monitor-C is not
	repo, db := setupRepo(t)
	ctx := t.Context()

	userID := createUser(t, db)
	monA := createMonitor(t, db, userID, "monitor-A")
	monB := createMonitor(t, db, userID, "monitor-B")
	_ = createMonitor(t, db, userID, "monitor-C") // untagged

	tagAlpha := createTag(t, db, "alpha")
	tagBeta := createTag(t, db, "beta")

	linkMonitorTag(t, db, monA, tagAlpha)
	linkMonitorTag(t, db, monB, tagBeta)

	// When: we query by both tag IDs
	results, err := repo.FindByTagIDs(ctx, []string{tagAlpha, tagBeta})

	// Then: only the two tagged monitors are returned
	require.NoError(t, err)
	assert.Len(t, results, 2)

	ids := make([]string, len(results))
	for i, m := range results {
		ids[i] = m.ID
	}
	assert.Contains(t, ids, monA)
	assert.Contains(t, ids, monB)
}

// TestFindByTagIDs_DeduplicatesMultipleMatchingTags verifies that a monitor
// matching multiple queried tags appears only once in the result.
func TestFindByTagIDs_DeduplicatesMultipleMatchingTags(t *testing.T) {
	// Given: a single monitor with two tags
	repo, db := setupRepo(t)
	ctx := t.Context()

	userID := createUser(t, db)
	monID := createMonitor(t, db, userID, "multi-tagged")

	tag1 := createTag(t, db, "tag-one")
	tag2 := createTag(t, db, "tag-two")

	linkMonitorTag(t, db, monID, tag1)
	linkMonitorTag(t, db, monID, tag2)

	// When: we query by both tags that the same monitor has
	results, err := repo.FindByTagIDs(ctx, []string{tag1, tag2})

	// Then: the monitor appears exactly once (no duplicates)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, monID, results[0].ID)
}

// TestFindByTagIDs_EmptySliceReturnsNil verifies the short-circuit path:
// an empty tag ID slice returns nil, nil without hitting the database.
func TestFindByTagIDs_EmptySliceReturnsNil(t *testing.T) {
	// Given: a repository (data content is irrelevant)
	repo, _ := setupRepo(t)
	ctx := t.Context()

	// When: we pass an empty slice
	results, err := repo.FindByTagIDs(ctx, []string{})

	// Then: both return values are nil
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// TestGroupTagIDs_PersistsRoundTrip verifies that GroupTagIDs written during
// Create is readable via FindByID with the value intact.
func TestGroupTagIDs_PersistsRoundTrip(t *testing.T) {
	repo, db := setupRepo(t)
	ctx := t.Context()

	userID := createUser(t, db)
	tagID := uuid.New().String()

	m := &monitor.Monitor{
		Name:        "My Group",
		Type:        "group",
		Active:      true,
		UserID:      userID,
		Interval:    60,
		Timeout:     48,
		GroupTagIDs: `["` + tagID + `"]`,
	}

	id, err := repo.Create(ctx, m)
	require.NoError(t, err)

	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, `["`+tagID+`"]`, loaded.GroupTagIDs)
}

// TestGroupTagIDs_UpdatePersists verifies that updating GroupTagIDs writes
// the new value to the database.
func TestGroupTagIDs_UpdatePersists(t *testing.T) {
	repo, db := setupRepo(t)
	ctx := t.Context()

	userID := createUser(t, db)

	m := &monitor.Monitor{
		Name:     "My Group",
		Type:     "group",
		Active:   true,
		UserID:   userID,
		Interval: 60,
		Timeout:  48,
	}

	id, err := repo.Create(ctx, m)
	require.NoError(t, err)

	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "", loaded.GroupTagIDs)

	tagID := uuid.New().String()
	loaded.GroupTagIDs = `["` + tagID + `"]`
	require.NoError(t, repo.Update(ctx, loaded))

	reloaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, `["`+tagID+`"]`, reloaded.GroupTagIDs)
}
