package statuspage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
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

func (r *sqliteRepo) FindAll(ctx context.Context) ([]*StatusPage, error) {
	pages, err := models.StatusPages.Query().All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying status pages: %w", err)
	}

	result := make([]*StatusPage, len(pages))
	for i, p := range pages {
		result[i] = statusPageFromModel(p)
	}
	return result, nil
}

func (r *sqliteRepo) FindBySlug(ctx context.Context, slug string) (*StatusPage, error) {
	p, err := models.StatusPages.Query(
		sm.Where(models.StatusPages.Columns.Slug.EQ(sqlite.Arg(slug))),
	).One(ctx, r.db)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding status page by slug") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return statusPageFromModel(p), nil
}

func (r *sqliteRepo) Create(ctx context.Context, sp *StatusPage) (string, error) {
	sp.ID = uuid.New().String()

	_, err := models.StatusPages.Insert(&models.StatusPageSetter{
		ID:                    omit.From(sp.ID),
		Slug:                  omit.From(sp.Slug),
		Title:                 omit.From(sp.Title),
		Description:           omitnull.From(sp.Description),
		Icon:                  omitnull.From(sp.Icon),
		Theme:                 omit.From(sp.Theme),
		Published:             omit.From(sp.Published),
		ShowTags:              omit.From(sp.ShowTags),
		ShowPoweredBy:         omit.From(sp.ShowPoweredBy),
		ShowCertificateExpiry: omit.From(sp.ShowCertificateExpiry),
		CustomCSS:             omitnull.From(sp.CustomCSS),
		FooterText:            omitnull.From(sp.FooterText),
		GoogleAnalyticsTagID:  omitnull.From(sp.GoogleAnalyticsTagID),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating status page: %w", err)
	}
	return sp.ID, nil
}

func (r *sqliteRepo) Update(ctx context.Context, sp *StatusPage) error {
	existing, err := models.FindStatusPage(ctx, r.db, sp.ID)
	if err != nil {
		return errs.WrapNotFound(err, "finding status page for update") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	now := time.Now()
	if err := existing.Update(ctx, r.db, &models.StatusPageSetter{
		Slug:                  omit.From(sp.Slug),
		Title:                 omit.From(sp.Title),
		Description:           omitnull.From(sp.Description),
		Icon:                  omitnull.From(sp.Icon),
		Theme:                 omit.From(sp.Theme),
		Published:             omit.From(sp.Published),
		ShowTags:              omit.From(sp.ShowTags),
		ShowPoweredBy:         omit.From(sp.ShowPoweredBy),
		ShowCertificateExpiry: omit.From(sp.ShowCertificateExpiry),
		CustomCSS:             omitnull.From(sp.CustomCSS),
		FooterText:            omitnull.From(sp.FooterText),
		GoogleAnalyticsTagID:  omitnull.From(sp.GoogleAnalyticsTagID),
		ModifiedDate:          omitnull.From(now),
	}); err != nil {
		return fmt.Errorf("updating status page: %w", err)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, id string) error {
	sp, err := models.FindStatusPage(ctx, r.db, id)
	if err != nil {
		return errs.WrapNotFound(err, "finding status page for delete") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := sp.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting status page: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetGroups(ctx context.Context, statusPageID string) ([]*Group, error) {
	groups, err := models.Groups.Query(
		sm.Where(models.Groups.Columns.StatusPageID.EQ(sqlite.Arg(statusPageID))),
		sm.OrderBy(models.Groups.Columns.Weight).Asc(),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying groups for status page: %w", err)
	}

	result := make([]*Group, len(groups))
	for i, g := range groups {
		result[i] = &Group{
			ID:           g.ID,
			Name:         g.Name,
			Weight:       g.Weight,
			StatusPageID: g.StatusPageID.GetOrZero(),
			TagIDs:       g.TagIdsJSON.GetOrZero(),
		}
	}
	return result, nil
}

func (r *sqliteRepo) SaveGroups(ctx context.Context, statusPageID string, groups []*Group) error {
	existing, err := models.Groups.Query(
		sm.Where(models.Groups.Columns.StatusPageID.EQ(sqlite.Arg(statusPageID))),
	).All(ctx, r.db)
	if err != nil {
		return fmt.Errorf("querying existing groups: %w", err)
	}

	if err := existing.DeleteAll(ctx, r.db); err != nil {
		return fmt.Errorf("deleting existing groups: %w", err)
	}

	for _, g := range groups {
		if g.ID == "" {
			g.ID = uuid.New().String()
		}
		_, insertErr := models.Groups.Insert(&models.GroupSetter{
			ID:           omit.From(g.ID),
			Name:         omit.From(g.Name),
			Weight:       omit.From(g.Weight),
			StatusPageID: omitnull.From(statusPageID),
			TagIdsJSON:   omitnull.From(g.TagIDs),
		}).One(ctx, r.db)
		if insertErr != nil {
			return fmt.Errorf("inserting group %q: %w", g.Name, insertErr)
		}
	}
	return nil
}

func (r *sqliteRepo) GetMonitorGroups(ctx context.Context, groupID string) ([]*MonitorGroup, error) {
	mgs, err := models.MonitorGroups.Query(
		sm.Where(models.MonitorGroups.Columns.GroupID.EQ(sqlite.Arg(groupID))),
		sm.OrderBy(models.MonitorGroups.Columns.Weight).Asc(),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying monitor groups: %w", err)
	}

	result := make([]*MonitorGroup, len(mgs))
	for i, mg := range mgs {
		result[i] = &MonitorGroup{
			MonitorID: mg.MonitorID,
			GroupID:   mg.GroupID,
			Weight:    mg.Weight,
		}
	}
	return result, nil
}

func (r *sqliteRepo) SaveMonitorGroups(ctx context.Context, groupID string, monitorGroups []*MonitorGroup) error {
	existing, err := models.MonitorGroups.Query(
		sm.Where(models.MonitorGroups.Columns.GroupID.EQ(sqlite.Arg(groupID))),
	).All(ctx, r.db)
	if err != nil {
		return fmt.Errorf("querying existing monitor groups: %w", err)
	}

	if err := existing.DeleteAll(ctx, r.db); err != nil {
		return fmt.Errorf("deleting existing monitor groups: %w", err)
	}

	for _, mg := range monitorGroups {
		_, insertErr := models.MonitorGroups.Insert(&models.MonitorGroupSetter{
			ID:        omit.From(uuid.New().String()),
			MonitorID: omit.From(mg.MonitorID),
			GroupID:   omit.From(groupID),
			Weight:    omit.From(mg.Weight),
		}).One(ctx, r.db)
		if insertErr != nil {
			return fmt.Errorf("inserting monitor group: %w", insertErr)
		}
	}
	return nil
}

func (r *sqliteRepo) FindIncidentsByStatusPage(ctx context.Context, statusPageID string) ([]*Incident, error) {
	incidents, err := models.Incidents.Query(
		sm.Where(models.Incidents.Columns.StatusPageID.EQ(sqlite.Arg(statusPageID))),
		sm.OrderBy(models.Incidents.Columns.CreatedDate).Desc(),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("querying incidents: %w", err)
	}

	result := make([]*Incident, len(incidents))
	for i, inc := range incidents {
		result[i] = incidentFromModel(inc)
	}
	return result, nil
}

func (r *sqliteRepo) FindIncidentByID(ctx context.Context, id string) (*Incident, error) {
	inc, err := models.FindIncident(ctx, r.db, id)
	if err != nil {
		return nil, errs.WrapNotFound(err, "finding incident") //nolint:wrapcheck // WrapNotFound handles wrapping
	}
	return incidentFromModel(inc), nil
}

func (r *sqliteRepo) CreateIncident(ctx context.Context, inc *Incident) (string, error) {
	inc.ID = uuid.New().String()

	_, err := models.Incidents.Insert(&models.IncidentSetter{
		ID:           omit.From(inc.ID),
		Title:        omit.From(inc.Title),
		Content:      omit.From(inc.Content),
		Style:        omit.From(inc.Style),
		Pin:          omit.From(inc.Pin),
		Active:       omit.From(inc.Active),
		StatusPageID: omitnull.From(inc.StatusPageID),
	}).One(ctx, r.db)
	if err != nil {
		return "", fmt.Errorf("creating incident: %w", err)
	}
	return inc.ID, nil
}

func (r *sqliteRepo) UpdateIncident(ctx context.Context, inc *Incident) error {
	existing, err := models.FindIncident(ctx, r.db, inc.ID)
	if err != nil {
		return errs.WrapNotFound(err, "finding incident for update") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	setter := &models.IncidentSetter{
		Title:   omit.From(inc.Title),
		Content: omit.From(inc.Content),
		Style:   omit.From(inc.Style),
		Pin:     omit.From(inc.Pin),
		Active:  omit.From(inc.Active),
	}
	if inc.LastUpdatedDate != nil {
		setter.LastUpdatedDate = omitnull.From(*inc.LastUpdatedDate)
	}

	if err := existing.Update(ctx, r.db, setter); err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}
	return nil
}

func (r *sqliteRepo) DeleteIncident(ctx context.Context, id string) error {
	inc, err := models.FindIncident(ctx, r.db, id)
	if err != nil {
		return errs.WrapNotFound(err, "finding incident for delete") //nolint:wrapcheck // WrapNotFound handles wrapping
	}

	if err := inc.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("deleting incident: %w", err)
	}
	return nil
}

func statusPageFromModel(m *models.StatusPage) *StatusPage {
	return &StatusPage{
		ID:                    m.ID,
		Slug:                  m.Slug,
		Title:                 m.Title,
		Description:           m.Description.GetOrZero(),
		Icon:                  m.Icon.GetOrZero(),
		Theme:                 m.Theme,
		Published:             m.Published,
		ShowTags:              m.ShowTags,
		ShowPoweredBy:         m.ShowPoweredBy,
		ShowCertificateExpiry: m.ShowCertificateExpiry,
		CustomCSS:             m.CustomCSS.GetOrZero(),
		FooterText:            m.FooterText.GetOrZero(),
		GoogleAnalyticsTagID:  m.GoogleAnalyticsTagID.GetOrZero(),
	}
}

func incidentFromModel(m *models.Incident) *Incident {
	inc := &Incident{
		ID:           m.ID,
		Title:        m.Title,
		Content:      m.Content,
		Style:        m.Style,
		Pin:          m.Pin,
		Active:       m.Active,
		StatusPageID: m.StatusPageID.GetOrZero(),
		CreatedDate:  m.CreatedDate,
	}
	if t, ok := m.LastUpdatedDate.Get(); ok {
		inc.LastUpdatedDate = &t
	}
	return inc
}
