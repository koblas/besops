package tag

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/google/uuid"
	"github.com/koblas/besops/lib/errs"
	"github.com/koblas/besops/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type sqliteRepo struct {
	db bob.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{db: bob.NewDB(db)}
}

func (r *sqliteRepo) FindAll(ctx context.Context) ([]*Tag, error) {
	tags, err := models.Tags.Query().All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}

	result := make([]*Tag, len(tags))
	for i, t := range tags {
		result[i] = tagFromModel(t)
	}
	return result, nil
}

func (r *sqliteRepo) FindByID(ctx context.Context, id string) (*Tag, error) {
	t, err := models.FindTag(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding tag") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return tagFromModel(t), nil
}

func (r *sqliteRepo) Create(ctx context.Context, t *Tag) (string, error) {
	t.ID = uuid.New().String()

	_, err := models.Tags.Insert(&models.TagSetter{
		ID:    omit.From(t.ID),
		Name:  omit.From(t.Name),
		Color: omit.From(t.Color),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("inserting tag: %w", err)
	}
	return t.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, t *Tag) error {
	tag, err := models.FindTag(ctx, r.db, t.ID)
	if err != nil {
		return fmt.Errorf("finding tag for update: %w", err)
	}

	err = tag.Update(ctx, r.db, &models.TagSetter{
		Name:  omit.From(t.Name),
		Color: omit.From(t.Color),
	})
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	tag, err := models.FindTag(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("finding tag for delete: %w", err)
	}

	err = tag.Delete(ctx, r.db)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetForMonitor(ctx context.Context, monitorID string) ([]*MonitorTag, error) {
	mts, err := models.MonitorTags.Query(
		sm.Where(models.MonitorTags.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitor tags: %w", err)
	}

	result := make([]*MonitorTag, len(mts))
	for i, mt := range mts {
		result[i] = monitorTagFromModel(mt)
	}
	return result, nil
}

func (r *sqliteRepo) AddToMonitor(ctx context.Context, monitorID, tagID string, value string) error {
	id := uuid.New().String()

	_, err := models.MonitorTags.Insert(&models.MonitorTagSetter{
		ID:        omit.From(id),
		MonitorID: omit.From(monitorID),
		TagID:     omit.From(tagID),
	}).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("adding tag to monitor: %w", err)
	}
	return nil
}

func (r *sqliteRepo) RemoveFromMonitor(ctx context.Context, monitorID, tagID string) error {
	mts, err := models.MonitorTags.Query(
		sm.Where(models.MonitorTags.Columns.MonitorID.EQ(sqlite.Arg(monitorID))),
		sm.Where(models.MonitorTags.Columns.TagID.EQ(sqlite.Arg(tagID))),
	).All(ctx, r.db)
	if err != nil {
		return fmt.Errorf("finding monitor tag for removal: %w", err)
	}

	if err := mts.DeleteAll(ctx, r.db); err != nil {
		return fmt.Errorf("removing tag from monitor: %w", err)
	}
	return nil
}

func tagFromModel(m *models.Tag) *Tag {
	return &Tag{
		ID:    m.ID,
		Name:  m.Name,
		Color: m.Color,
	}
}

func monitorTagFromModel(m *models.MonitorTag) *MonitorTag {
	return &MonitorTag{
		ID:        m.ID,
		MonitorID: m.MonitorID,
		TagID:     m.TagID,
		Value:     m.Value.GetOrZero(),
	}
}
