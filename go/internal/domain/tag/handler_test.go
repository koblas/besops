package tag_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/database"
	"github.com/koblas/besops/internal/domain/tag"
	"github.com/stretchr/testify/require"
)

func setupHandler(t *testing.T) (*tag.Handler, *handlerFixture) {
	t.Helper()
	dir := t.TempDir()
	dbURL := "sqlite://" + filepath.Join(dir, "test.db")

	db, err := database.Open(dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, database.Migrate(db, dbURL))

	repo := tag.NewRepository(db)
	handler := tag.NewHandler(repo)

	userID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO "user" (id, username, password) VALUES (?, 'testuser', 'hashed')`, userID)
	require.NoError(t, err)

	monitorID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO monitor (id, name, type, active, user_id) VALUES (?, 'Handler Test Monitor', 'http', 1, ?)`, monitorID, userID)
	require.NoError(t, err)

	return handler, &handlerFixture{repo: repo, monitorID: monitorID}
}

type handlerFixture struct {
	repo      tag.Repository
	monitorID string
}

func TestHandler_CreateTag(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	res, err := handler.CreateTag(ctx, &oas.TagInput{
		Name:  "new-tag",
		Color: "#ff0000",
	})
	require.NoError(t, err)
	result := res.(*oas.Tag)
	require.Equal(t, "new-tag", result.Name)
	require.Equal(t, "#ff0000", result.Color)
	require.NotEqual(t, uuid.Nil, result.ID)
}

func TestHandler_ListTags_Empty(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	res, err := handler.ListTags(ctx)
	require.NoError(t, err)
	tags := res.(*oas.ListTagsOKApplicationJSON)
	require.Empty(t, *tags)
}

func TestHandler_ListTags_ReturnsTags(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	_, err := handler.CreateTag(ctx, &oas.TagInput{Name: "a", Color: "#a"})
	require.NoError(t, err)
	_, err = handler.CreateTag(ctx, &oas.TagInput{Name: "b", Color: "#b"})
	require.NoError(t, err)

	res, err := handler.ListTags(ctx)
	require.NoError(t, err)
	tags := res.(*oas.ListTagsOKApplicationJSON)
	require.Len(t, *tags, 2)
}

func TestHandler_UpdateTag(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "old", Color: "#111"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	updateRes, err := handler.UpdateTag(ctx, &oas.TagInput{Name: "new", Color: "#222"}, oas.UpdateTagParams{
		TagId: created.ID,
	})
	require.NoError(t, err)
	updated := updateRes.(*oas.Tag)
	require.Equal(t, "new", updated.Name)
	require.Equal(t, "#222", updated.Color)
}

func TestHandler_DeleteTag(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "delete-me", Color: "#000"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	_, err = handler.DeleteTag(ctx, oas.DeleteTagParams{TagId: created.ID})
	require.NoError(t, err)

	// Verify it's gone
	res, err := handler.ListTags(ctx)
	require.NoError(t, err)
	tags := res.(*oas.ListTagsOKApplicationJSON)
	require.Empty(t, *tags)
}

func TestHandler_AddMonitorTag(t *testing.T) {
	ctx := context.Background()
	handler, fix := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "assign-me", Color: "#abc"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	monitorUUID, _ := uuid.Parse(fix.monitorID)
	_, err = handler.AddMonitorTag(ctx, &oas.AddMonitorTagReq{
		TagId: created.ID,
	}, oas.AddMonitorTagParams{
		MonitorId: monitorUUID,
	})
	require.NoError(t, err)

	// Verify via repository
	mts, err := fix.repo.GetForMonitor(ctx, fix.monitorID)
	require.NoError(t, err)
	require.Len(t, mts, 1)
	require.Equal(t, created.ID.String(), mts[0].TagID)
}

func TestHandler_AddMonitorTag_InvalidMonitor(t *testing.T) {
	ctx := context.Background()
	handler, _ := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "orphan", Color: "#000"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	fakeMonitorID := uuid.New()
	_, err = handler.AddMonitorTag(ctx, &oas.AddMonitorTagReq{
		TagId: created.ID,
	}, oas.AddMonitorTagParams{
		MonitorId: fakeMonitorID,
	})
	require.Error(t, err, "should fail for nonexistent monitor")
}

func TestHandler_AddMonitorTag_InvalidTag(t *testing.T) {
	ctx := context.Background()
	handler, fix := setupHandler(t)

	monitorUUID, _ := uuid.Parse(fix.monitorID)
	_, err := handler.AddMonitorTag(ctx, &oas.AddMonitorTagReq{
		TagId: uuid.New(),
	}, oas.AddMonitorTagParams{
		MonitorId: monitorUUID,
	})
	require.Error(t, err, "should fail for nonexistent tag")
}

func TestHandler_DeleteMonitorTag(t *testing.T) {
	ctx := context.Background()
	handler, fix := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "remove-me", Color: "#f00"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	monitorUUID, _ := uuid.Parse(fix.monitorID)
	_, err = handler.AddMonitorTag(ctx, &oas.AddMonitorTagReq{
		TagId: created.ID,
	}, oas.AddMonitorTagParams{
		MonitorId: monitorUUID,
	})
	require.NoError(t, err)

	_, err = handler.DeleteMonitorTag(ctx, oas.DeleteMonitorTagParams{
		MonitorId: monitorUUID,
		TagId:     created.ID,
	})
	require.NoError(t, err)

	mts, err := fix.repo.GetForMonitor(ctx, fix.monitorID)
	require.NoError(t, err)
	require.Empty(t, mts)
}

func TestHandler_DeleteMonitorTag_NotAssigned(t *testing.T) {
	ctx := context.Background()
	handler, fix := setupHandler(t)

	createRes, err := handler.CreateTag(ctx, &oas.TagInput{Name: "never-assigned", Color: "#000"})
	require.NoError(t, err)
	created := createRes.(*oas.Tag)

	monitorUUID, _ := uuid.Parse(fix.monitorID)
	// Should not error when removing a tag that isn't assigned
	_, err = handler.DeleteMonitorTag(ctx, oas.DeleteMonitorTagParams{
		MonitorId: monitorUUID,
		TagId:     created.ID,
	})
	require.NoError(t, err)
}
