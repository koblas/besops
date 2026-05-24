package statuspage

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

func TestParseTagIDs_Empty(t *testing.T) {
	result := parseTagIDs("")
	assert.Nil(t, result)
}

func TestParseTagIDs_ValidJSON(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	input := `["` + id1.String() + `","` + id2.String() + `"]`

	result := parseTagIDs(input)
	require.Len(t, result, 2)
	assert.Equal(t, id1, result[0])
	assert.Equal(t, id2, result[1])
}

func TestParseTagIDs_InvalidJSON(t *testing.T) {
	result := parseTagIDs("not json")
	assert.Nil(t, result)
}

func TestSerializeTagIDs_Empty(t *testing.T) {
	result := serializeTagIDs(nil)
	assert.Equal(t, "", result)
}

func TestSerializeTagIDs_RoundTrip(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New()}

	serialized := serializeTagIDs(ids)
	parsed := parseTagIDs(serialized)

	require.Len(t, parsed, 2)
	assert.Equal(t, ids[0], parsed[0])
	assert.Equal(t, ids[1], parsed[1])
}

func TestParseTagIDStrings_Empty(t *testing.T) {
	result := parseTagIDStrings("")
	assert.Nil(t, result)
}

func TestParseTagIDStrings_ValidJSON(t *testing.T) {
	result := parseTagIDStrings(`["abc","def"]`)
	assert.Equal(t, []string{"abc", "def"}, result)
}

type mockMonitorResolver struct {
	ids []string
}

func (m *mockMonitorResolver) FindIDsByTagIDs(_ context.Context, _ []string) ([]string, error) {
	return m.ids, nil
}

func TestLoadGroups_IncludesTagIds(t *testing.T) {
	tagID := uuid.New()
	repo := &mockRepo{
		groups: []*Group{
			{ID: uuid.New().String(), Name: "Web", TagIDs: `["` + tagID.String() + `"]`},
		},
		monitorGroups: map[string][]*MonitorGroup{},
	}

	h := NewHandler(repo, nil, nil)
	groups, err := h.loadGroups(t.Context(), "sp-1")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Len(t, groups[0].TagIds, 1)
	assert.Equal(t, tagID, groups[0].TagIds[0])
}

func TestSaveGroups_PersistsTagIds(t *testing.T) {
	tagID := uuid.New()
	repo := &mockRepo{
		monitorGroups: map[string][]*MonitorGroup{},
	}

	h := NewHandler(repo, nil, nil)
	err := h.saveGroups(t.Context(), "sp-1", []oas.StatusPageGroup{
		{Name: "Infra", TagIds: []uuid.UUID{tagID}},
	})
	require.NoError(t, err)
	require.Len(t, repo.savedGroups, 1)
	assert.Equal(t, `["`+tagID.String()+`"]`, repo.savedGroups[0].TagIDs)
}

type mockHbReader struct {
	beats map[string][]*Heartbeat
}

func (m *mockHbReader) GetByMonitorPaged(_ context.Context, monitorID string, _, _ int) ([]*Heartbeat, error) {
	return m.beats[monitorID], nil
}

func (m *mockHbReader) GetUptime(_ context.Context, _ string, _ int) (float64, error) {
	return 0.99, nil
}

// Given a status page with a group that has both explicit monitors and tag-based monitors,
// When GetStatusPageHeartbeats is called,
// Then heartbeats are returned for all monitors (union, deduplicated).
func TestGetStatusPageHeartbeats_UnionsExplicitAndTagMonitors(t *testing.T) {
	groupID := uuid.New().String()
	explicitMon := uuid.New().String()
	taggedMon := uuid.New().String()
	sharedMon := uuid.New().String()

	repo := &mockRepo{
		statusPage: &StatusPage{ID: "sp-1", Slug: "prod"},
		groups: []*Group{
			{ID: groupID, Name: "Prod", TagIDs: `["tag-1"]`},
		},
		monitorGroups: map[string][]*MonitorGroup{
			groupID: {
				{MonitorID: explicitMon, GroupID: groupID},
				{MonitorID: sharedMon, GroupID: groupID},
			},
		},
	}

	hbReader := &mockHbReader{
		beats: map[string][]*Heartbeat{
			explicitMon: {{ID: "hb-1", MonitorID: explicitMon, Status: 1}},
			taggedMon:   {{ID: "hb-2", MonitorID: taggedMon, Status: 1}},
			sharedMon:   {{ID: "hb-3", MonitorID: sharedMon, Status: 1}},
		},
	}

	resolver := &mockMonitorResolver{ids: []string{taggedMon, sharedMon}}

	h := NewHandler(repo, hbReader, nil, WithMonitorResolver(resolver))

	res, err := h.GetStatusPageHeartbeats(t.Context(), oas.GetStatusPageHeartbeatsParams{Slug: "prod"})
	require.NoError(t, err)
	result := res.(*oas.GetStatusPageHeartbeatsOK)

	hbList := result.HeartbeatList
	// Should have all 3 unique monitors (explicit + tagged, with sharedMon deduplicated)
	assert.Len(t, hbList, 3)
	hbIDs := make(map[string]bool)
	for _, item := range hbList {
		hbIDs[item.MonitorId.String()] = true
	}
	assert.True(t, hbIDs[explicitMon], "explicit monitor should be in heartbeat list")
	assert.True(t, hbIDs[taggedMon], "tag-resolved monitor should be in heartbeat list")
	assert.True(t, hbIDs[sharedMon], "shared monitor should appear once (deduplicated)")
}

// mockRepo implements Repository for testing.
type mockRepo struct {
	statusPage    *StatusPage
	groups        []*Group
	savedGroups   []*Group
	monitorGroups map[string][]*MonitorGroup
}

func (m *mockRepo) FindAll(_ context.Context) ([]*StatusPage, error) { return nil, nil }
func (m *mockRepo) FindBySlug(_ context.Context, _ string) (*StatusPage, error) {
	if m.statusPage != nil {
		return m.statusPage, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockRepo) Create(_ context.Context, _ *StatusPage) (string, error) { return "", nil }
func (m *mockRepo) Update(_ context.Context, _ *StatusPage) error           { return nil }
func (m *mockRepo) Delete(_ context.Context, _ string) error                { return nil }
func (m *mockRepo) GetGroups(_ context.Context, _ string) ([]*Group, error) { return m.groups, nil }
func (m *mockRepo) SaveGroups(_ context.Context, _ string, groups []*Group) error {
	m.savedGroups = groups
	return nil
}
func (m *mockRepo) GetMonitorGroups(_ context.Context, groupID string) ([]*MonitorGroup, error) {
	return m.monitorGroups[groupID], nil
}
func (m *mockRepo) SaveMonitorGroups(_ context.Context, _ string, _ []*MonitorGroup) error {
	return nil
}
func (m *mockRepo) FindIncidentsByStatusPage(_ context.Context, _ string) ([]*Incident, error) {
	return nil, nil
}
func (m *mockRepo) FindIncidentByID(_ context.Context, _ string) (*Incident, error) { return nil, nil }
func (m *mockRepo) CreateIncident(_ context.Context, _ *Incident) (string, error)   { return "", nil }
func (m *mockRepo) UpdateIncident(_ context.Context, _ *Incident) error             { return nil }
func (m *mockRepo) DeleteIncident(_ context.Context, _ string) error                { return nil }
