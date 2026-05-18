package statuspage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
	"github.com/koblas/besops/internal/domain/heartbeat"
)

var (
	_ oas.StatusPageHandler = (*Handler)(nil)
	_ oas.IncidentHandler   = (*Handler)(nil)
)

// HeartbeatReader provides access to heartbeat data for the status page.
type HeartbeatReader interface {
	GetByMonitorPaged(ctx context.Context, monitorID string, offset, limit int) ([]*Heartbeat, error)
	GetUptime(ctx context.Context, monitorID string, hours int) (float64, error)
}

// Heartbeat is used locally to avoid importing the heartbeat domain package.
type Heartbeat = heartbeat.Heartbeat

// MonitorNameResolver provides monitor names without requiring full auth.
type MonitorNameResolver interface {
	FindNameByID(ctx context.Context, id string) (string, error)
}

// MonitorResolver resolves tag IDs to monitor IDs for status page groups.
type MonitorResolver interface {
	FindIDsByTagIDs(ctx context.Context, tagIDs []string) ([]string, error)
}

type Handler struct {
	repo            Repository
	hbReader        HeartbeatReader
	monitorNam      MonitorNameResolver
	monitorResolver MonitorResolver
}

func NewHandler(repo Repository, hbReader HeartbeatReader, monitorNam MonitorNameResolver, opts ...HandlerOption) *Handler {
	h := &Handler{repo: repo, hbReader: hbReader, monitorNam: monitorNam}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type HandlerOption func(*Handler)

func WithMonitorResolver(r MonitorResolver) HandlerOption {
	return func(h *Handler) { h.monitorResolver = r }
}

func (h *Handler) ListStatusPages(ctx context.Context) ([]oas.StatusPage, error) {
	pages, err := h.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing status pages: %w", err)
	}

	result := make([]oas.StatusPage, len(pages))
	for i, sp := range pages {
		groups, groupErr := h.loadGroups(ctx, sp.ID)
		if groupErr != nil {
			return nil, groupErr
		}
		result[i] = statusPageToOAS(sp, groups)
	}
	return result, nil
}

func (h *Handler) GetStatusPage(ctx context.Context, params oas.GetStatusPageParams) (oas.GetStatusPageRes, error) {
	sp, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return &oas.ErrorResponse{Error: "status page not found"}, nil //nolint:nilerr
	}

	groups, groupErr := h.loadGroups(ctx, sp.ID)
	if groupErr != nil {
		return nil, groupErr
	}

	result := statusPageToOAS(sp, groups)
	return &result, nil
}

func (h *Handler) CreateStatusPage(ctx context.Context, req *oas.StatusPageInput) (*oas.CreateStatusPageCreated, error) {
	sp := statusPageFromInput(req)

	id, err := h.repo.Create(ctx, sp)
	if err != nil {
		return nil, fmt.Errorf("creating status page: %w", err)
	}

	if saveErr := h.saveGroups(ctx, id, req.Groups); saveErr != nil {
		return nil, saveErr
	}

	return &oas.CreateStatusPageCreated{Slug: req.Slug}, nil
}

func (h *Handler) UpdateStatusPage(ctx context.Context, req *oas.StatusPageInput, params oas.UpdateStatusPageParams) (*oas.StatusPage, error) {
	existing, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("finding status page: %w", err)
	}

	sp := statusPageFromInput(req)
	sp.ID = existing.ID

	if updateErr := h.repo.Update(ctx, sp); updateErr != nil {
		return nil, fmt.Errorf("updating status page: %w", updateErr)
	}

	if saveErr := h.saveGroups(ctx, sp.ID, req.Groups); saveErr != nil {
		return nil, saveErr
	}

	groups, groupErr := h.loadGroups(ctx, sp.ID)
	if groupErr != nil {
		return nil, groupErr
	}

	result := statusPageToOAS(sp, groups)
	return &result, nil
}

func (h *Handler) DeleteStatusPage(ctx context.Context, params oas.DeleteStatusPageParams) error {
	sp, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return fmt.Errorf("finding status page for delete: %w", err)
	}

	if deleteErr := h.repo.Delete(ctx, sp.ID); deleteErr != nil {
		return fmt.Errorf("deleting status page: %w", deleteErr)
	}
	return nil
}

func (h *Handler) GetStatusPageHeartbeats(ctx context.Context, params oas.GetStatusPageHeartbeatsParams) (*oas.GetStatusPageHeartbeatsOK, error) {
	sp, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("finding status page: %w", err)
	}

	groups, err := h.repo.GetGroups(ctx, sp.ID)
	if err != nil {
		return nil, fmt.Errorf("loading groups: %w", err)
	}

	seen := map[string]struct{}{}
	var monitorIDs []string
	for _, g := range groups {
		mgs, mgErr := h.repo.GetMonitorGroups(ctx, g.ID)
		if mgErr != nil {
			return nil, fmt.Errorf("loading monitor groups: %w", mgErr)
		}
		for _, mg := range mgs {
			if _, ok := seen[mg.MonitorID]; !ok {
				seen[mg.MonitorID] = struct{}{}
				monitorIDs = append(monitorIDs, mg.MonitorID)
			}
		}

		tagIDs := parseTagIDStrings(g.TagIDs)
		if len(tagIDs) > 0 && h.monitorResolver != nil {
			resolved, resolveErr := h.monitorResolver.FindIDsByTagIDs(ctx, tagIDs)
			if resolveErr != nil {
				return nil, fmt.Errorf("resolving tag monitors: %w", resolveErr)
			}
			for _, mid := range resolved {
				if _, ok := seen[mid]; !ok {
					seen[mid] = struct{}{}
					monitorIDs = append(monitorIDs, mid)
				}
			}
		}
	}

	hbList := make(oas.GetStatusPageHeartbeatsOKHeartbeatList)
	uptimeList := make(oas.GetStatusPageHeartbeatsOKUptimeList)
	nameMap := make(oas.GetStatusPageHeartbeatsOKMonitorNames)

	for _, mid := range monitorIDs {
		beats, hbErr := h.hbReader.GetByMonitorPaged(ctx, mid, 0, 50)
		if hbErr != nil {
			continue
		}

		oasBeats := make([]oas.Heartbeat, len(beats))
		for i, hb := range beats {
			oasBeats[i] = hbToOAS(hb)
		}
		hbList[mid] = oasBeats

		uptime, uptimeErr := h.hbReader.GetUptime(ctx, mid, 24)
		if uptimeErr == nil {
			uptimeList[mid+"_24"] = uptime
		}

		if h.monitorNam != nil {
			name, nameErr := h.monitorNam.FindNameByID(ctx, mid)
			if nameErr == nil {
				nameMap[mid] = name
			}
		}
	}

	return &oas.GetStatusPageHeartbeatsOK{
		HeartbeatList: oas.NewOptGetStatusPageHeartbeatsOKHeartbeatList(hbList),
		UptimeList:    oas.NewOptGetStatusPageHeartbeatsOKUptimeList(uptimeList),
		MonitorNames:  oas.NewOptGetStatusPageHeartbeatsOKMonitorNames(nameMap),
	}, nil
}

func hbToOAS(hb *heartbeat.Heartbeat) oas.Heartbeat {
	result := oas.Heartbeat{
		ID:        oasutil.MustParseUUID(hb.ID),
		MonitorId: oasutil.MustParseUUID(hb.MonitorID),
		Status:    oas.HeartbeatStatus(hb.Status),
		Time:      time.Time(hb.Time),
		Important: oas.NewOptBool(hb.Important),
	}
	if hb.Msg != "" {
		result.Msg = oas.NewOptString(hb.Msg)
	}
	if hb.Latency != nil {
		result.Latency = oas.NewOptInt64(*hb.Latency)
	}
	if hb.Duration > 0 {
		result.Duration = oas.NewOptInt64(hb.Duration)
	}
	return result
}

func (h *Handler) GetStatusPageBadge(_ context.Context, _ oas.GetStatusPageBadgeParams) (oas.GetStatusPageBadgeOK, error) {
	return oas.GetStatusPageBadgeOK{Data: strings.NewReader("")}, nil
}

func (h *Handler) GetStatusPageEventStream(_ context.Context, _ oas.GetStatusPageEventStreamParams) (oas.GetStatusPageEventStreamOK, error) {
	return oas.GetStatusPageEventStreamOK{Data: strings.NewReader("")}, nil
}

// --- Incidents ---

func (h *Handler) ListIncidents(ctx context.Context, params oas.ListIncidentsParams) (*oas.ListIncidentsOK, error) {
	sp, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("finding status page: %w", err)
	}

	incidents, err := h.repo.FindIncidentsByStatusPage(ctx, sp.ID)
	if err != nil {
		return nil, fmt.Errorf("listing incidents: %w", err)
	}

	oasIncidents := make([]oas.Incident, len(incidents))
	for i, inc := range incidents {
		oasIncidents[i] = incidentToOAS(inc)
	}

	return &oas.ListIncidentsOK{Incidents: oasIncidents}, nil
}

func (h *Handler) CreateIncident(ctx context.Context, req *oas.IncidentInput, params oas.CreateIncidentParams) (*oas.Incident, error) {
	sp, err := h.repo.FindBySlug(ctx, params.Slug)
	if err != nil {
		return nil, fmt.Errorf("finding status page: %w", err)
	}

	inc := &Incident{
		Title:        req.Title,
		Content:      req.Content,
		Style:        string(req.Style),
		Pin:          oasutil.OptBoolValue(req.Pin, true),
		Active:       true,
		StatusPageID: sp.ID,
	}

	if _, createErr := h.repo.CreateIncident(ctx, inc); createErr != nil {
		return nil, fmt.Errorf("creating incident: %w", createErr)
	}

	result := incidentToOAS(inc)
	return &result, nil
}

func (h *Handler) UpdateIncident(ctx context.Context, req *oas.IncidentInput, params oas.UpdateIncidentParams) (*oas.Incident, error) {
	existing, err := h.repo.FindIncidentByID(ctx, params.IncidentId.String())
	if err != nil {
		return nil, fmt.Errorf("finding incident: %w", err)
	}

	now := time.Now()
	existing.Title = req.Title
	existing.Content = req.Content
	existing.Style = string(req.Style)
	existing.Pin = oasutil.OptBoolValue(req.Pin, existing.Pin)
	existing.LastUpdatedDate = &now

	if updateErr := h.repo.UpdateIncident(ctx, existing); updateErr != nil {
		return nil, fmt.Errorf("updating incident: %w", updateErr)
	}

	result := incidentToOAS(existing)
	return &result, nil
}

func (h *Handler) DeleteIncident(ctx context.Context, params oas.DeleteIncidentParams) error {
	if err := h.repo.DeleteIncident(ctx, params.IncidentId.String()); err != nil {
		return fmt.Errorf("deleting incident: %w", err)
	}
	return nil
}

func (h *Handler) ResolveIncident(ctx context.Context, params oas.ResolveIncidentParams) (*oas.Incident, error) {
	inc, err := h.repo.FindIncidentByID(ctx, params.IncidentId.String())
	if err != nil {
		return nil, fmt.Errorf("finding incident: %w", err)
	}

	now := time.Now()
	inc.Active = false
	inc.LastUpdatedDate = &now

	if updateErr := h.repo.UpdateIncident(ctx, inc); updateErr != nil {
		return nil, fmt.Errorf("resolving incident: %w", updateErr)
	}

	result := incidentToOAS(inc)
	return &result, nil
}

func (h *Handler) UnpinIncident(ctx context.Context, params oas.UnpinIncidentParams) (*oas.MessageResponse, error) {
	inc, err := h.repo.FindIncidentByID(ctx, params.IncidentId.String())
	if err != nil {
		return nil, fmt.Errorf("finding incident: %w", err)
	}

	now := time.Now()
	inc.Pin = false
	inc.LastUpdatedDate = &now

	if updateErr := h.repo.UpdateIncident(ctx, inc); updateErr != nil {
		return nil, fmt.Errorf("unpinning incident: %w", updateErr)
	}

	return &oas.MessageResponse{Message: "Incident unpinned"}, nil
}

// --- helpers ---

func (h *Handler) loadGroups(ctx context.Context, statusPageID string) ([]oas.StatusPageGroup, error) {
	groups, err := h.repo.GetGroups(ctx, statusPageID)
	if err != nil {
		return nil, fmt.Errorf("loading groups: %w", err)
	}

	result := make([]oas.StatusPageGroup, len(groups))
	for i, g := range groups {
		mgs, mgErr := h.repo.GetMonitorGroups(ctx, g.ID)
		if mgErr != nil {
			return nil, fmt.Errorf("loading monitor groups: %w", mgErr)
		}

		monitorIDs := make([]uuid.UUID, len(mgs))
		for j, mg := range mgs {
			monitorIDs[j] = uuid.MustParse(mg.MonitorID)
		}

		result[i] = oas.StatusPageGroup{
			ID:         oas.NewOptUUID(uuid.MustParse(g.ID)),
			Name:       g.Name,
			Weight:     oas.NewOptInt(int(g.Weight)),
			MonitorIds: monitorIDs,
			TagIds:     parseTagIDs(g.TagIDs),
		}
	}
	return result, nil
}

func (h *Handler) saveGroups(ctx context.Context, statusPageID string, oasGroups []oas.StatusPageGroup) error {
	groups := make([]*Group, len(oasGroups))
	for i, og := range oasGroups {
		id := ""
		if og.ID.IsSet() {
			id = og.ID.Value.String()
		}
		groups[i] = &Group{
			ID:           id,
			Name:         og.Name,
			Weight:       int64(oasutil.OptIntValue(og.Weight, 1000)),
			StatusPageID: statusPageID,
			TagIDs:       serializeTagIDs(og.TagIds),
		}
	}

	if err := h.repo.SaveGroups(ctx, statusPageID, groups); err != nil {
		return fmt.Errorf("saving groups: %w", err)
	}

	for i, g := range groups {
		if i >= len(oasGroups) {
			break
		}
		mgs := make([]*MonitorGroup, len(oasGroups[i].MonitorIds))
		for j, mid := range oasGroups[i].MonitorIds {
			mgs[j] = &MonitorGroup{
				MonitorID: mid.String(),
				GroupID:   g.ID,
				Weight:    int64(j),
			}
		}
		if mgErr := h.repo.SaveMonitorGroups(ctx, g.ID, mgs); mgErr != nil {
			return fmt.Errorf("saving monitor groups: %w", mgErr)
		}
	}
	return nil
}

func statusPageFromInput(req *oas.StatusPageInput) *StatusPage {
	theme := "auto"
	if req.Theme.IsSet() {
		theme = string(req.Theme.Value)
	}

	return &StatusPage{
		Slug:                  req.Slug,
		Title:                 req.Title,
		Description:           oasutil.OptStringValue(req.Description),
		Icon:                  oasutil.OptStringValue(req.Icon),
		Theme:                 theme,
		Published:             oasutil.OptBoolValue(req.Published, true),
		ShowTags:              oasutil.OptBoolValue(req.ShowTags, false),
		ShowPoweredBy:         oasutil.OptBoolValue(req.ShowPoweredBy, true),
		ShowCertificateExpiry: oasutil.OptBoolValue(req.ShowCertificateExpiry, false),
		CustomCSS:             oasutil.OptStringValue(req.CustomCss),
		FooterText:            oasutil.OptStringValue(req.FooterText),
		GoogleAnalyticsTagID:  oasutil.OptStringValue(req.GoogleAnalyticsId),
	}
}

func statusPageToOAS(sp *StatusPage, groups []oas.StatusPageGroup) oas.StatusPage {
	return oas.StatusPage{
		Slug:                  sp.Slug,
		Title:                 sp.Title,
		Description:           oasutil.PtrToOptString(sp.Description),
		Icon:                  oasutil.PtrToOptString(sp.Icon),
		Theme:                 optTheme(sp.Theme),
		Published:             oas.NewOptBool(sp.Published),
		ShowTags:              oas.NewOptBool(sp.ShowTags),
		ShowPoweredBy:         oas.NewOptBool(sp.ShowPoweredBy),
		ShowCertificateExpiry: oas.NewOptBool(sp.ShowCertificateExpiry),
		CustomCss:             oasutil.PtrToOptString(sp.CustomCSS),
		FooterText:            oasutil.PtrToOptString(sp.FooterText),
		GoogleAnalyticsId:     oasutil.PtrToOptString(sp.GoogleAnalyticsTagID),
		Groups:                groups,
	}
}

func optTheme(t string) oas.OptStatusPageTheme {
	if t == "" {
		return oas.OptStatusPageTheme{}
	}
	return oas.NewOptStatusPageTheme(oas.StatusPageTheme(t))
}

func parseTagIDStrings(s string) []string {
	if s == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

func parseTagIDs(s string) []uuid.UUID {
	if s == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		result = append(result, u)
	}
	return result
}

func serializeTagIDs(ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	data, _ := json.Marshal(strs)
	return string(data)
}

func incidentToOAS(inc *Incident) oas.Incident {
	result := oas.Incident{
		ID:       uuid.MustParse(inc.ID),
		Title:    inc.Title,
		Content:  inc.Content,
		Style:    oas.IncidentStyle(inc.Style),
		Pin:      oas.NewOptBool(inc.Pin),
		Resolved: oas.NewOptBool(!inc.Active),
	}
	if !inc.CreatedDate.IsZero() {
		result.CreatedDate = oas.NewOptDateTime(inc.CreatedDate)
	}
	if inc.LastUpdatedDate != nil {
		result.LastUpdatedDate = oas.NewOptDateTime(*inc.LastUpdatedDate)
	}
	return result
}
