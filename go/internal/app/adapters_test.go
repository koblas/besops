package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	domainnotification "github.com/koblas/besops/internal/domain/notification"
	"github.com/koblas/besops/internal/domain/tag"
)

// --- notificationRuleAdapter ---

type stubNotifRepo struct {
	notifications []*domainnotification.Notification
	err           error
}

func (s *stubNotifRepo) GetForMonitor(_ context.Context, _ string) ([]*domainnotification.Notification, error) {
	return s.notifications, s.err
}

func (s *stubNotifRepo) FindByID(_ context.Context, _ string) (*domainnotification.Notification, error) {
	return nil, nil
}
func (s *stubNotifRepo) FindByUserID(_ context.Context, _ string) ([]*domainnotification.Notification, error) {
	return nil, nil
}
func (s *stubNotifRepo) Create(_ context.Context, _ *domainnotification.Notification) (string, error) {
	return "", nil
}
func (s *stubNotifRepo) Update(_ context.Context, _ *domainnotification.Notification) error {
	return nil
}
func (s *stubNotifRepo) Delete(_ context.Context, _ string) error { return nil }

func TestNotificationRuleAdapter_ParsesConfig(t *testing.T) {
	repo := &stubNotifRepo{
		notifications: []*domainnotification.Notification{
			{
				ID:     "n1",
				Name:   "Slack Alert",
				Active: true,
				Config: `{"type":"slack","webhookUrl":"https://hooks.slack.com/xxx","channel":"#alerts"}`,
			},
		},
	}

	adapter := &notificationRuleAdapter{repo: repo}
	rules, err := adapter.GetRulesForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	assert.Equal(t, "n1", rules[0].ID)
	assert.Equal(t, "Slack Alert", rules[0].Name)
	assert.Equal(t, "slack", rules[0].Type)
	assert.True(t, rules[0].Active)
	assert.Equal(t, "https://hooks.slack.com/xxx", rules[0].Config["webhookUrl"])
	assert.Equal(t, "#alerts", rules[0].Config["channel"])
	// "type" should be removed from Config map since it's promoted to Rule.Type
	_, hasType := rules[0].Config["type"]
	assert.False(t, hasType, "type should be extracted from config map")
}

func TestNotificationRuleAdapter_SkipsMalformedJSON(t *testing.T) {
	repo := &stubNotifRepo{
		notifications: []*domainnotification.Notification{
			{ID: "good", Name: "Good", Active: true, Config: `{"type":"webhook","url":"http://example.com"}`},
			{ID: "bad", Name: "Bad", Active: true, Config: `not json at all`},
			{ID: "also-good", Name: "Also Good", Active: true, Config: `{"type":"telegram","chatId":"123"}`},
		},
	}

	adapter := &notificationRuleAdapter{repo: repo}
	rules, err := adapter.GetRulesForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	require.Len(t, rules, 2, "malformed config should be silently skipped")

	assert.Equal(t, "good", rules[0].ID)
	assert.Equal(t, "also-good", rules[1].ID)
}

func TestNotificationRuleAdapter_InactiveRulePreserved(t *testing.T) {
	repo := &stubNotifRepo{
		notifications: []*domainnotification.Notification{
			{ID: "n1", Name: "Disabled", Active: false, Config: `{"type":"smtp","to":"dev@example.com"}`},
		},
	}

	adapter := &notificationRuleAdapter{repo: repo}
	rules, err := adapter.GetRulesForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.False(t, rules[0].Active)
}

func TestNotificationRuleAdapter_EmptyNotifications(t *testing.T) {
	repo := &stubNotifRepo{notifications: nil}

	adapter := &notificationRuleAdapter{repo: repo}
	rules, err := adapter.GetRulesForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	assert.Empty(t, rules)
}

// --- tagProviderAdapter ---

type stubTagRepo struct {
	monitorTags []*tag.MonitorTag
	tags        map[string]*tag.Tag
}

func (s *stubTagRepo) GetForMonitor(_ context.Context, _ string) ([]*tag.MonitorTag, error) {
	return s.monitorTags, nil
}

func (s *stubTagRepo) FindByID(_ context.Context, id string) (*tag.Tag, error) {
	t, ok := s.tags[id]
	if !ok {
		return nil, assert.AnError
	}
	return t, nil
}

func (s *stubTagRepo) FindAll(_ context.Context) ([]*tag.Tag, error)          { return nil, nil }
func (s *stubTagRepo) Create(_ context.Context, _ *tag.Tag) (string, error)   { return "", nil }
func (s *stubTagRepo) Update(_ context.Context, _ *tag.Tag) error             { return nil }
func (s *stubTagRepo) Delete(_ context.Context, _ string) error               { return nil }
func (s *stubTagRepo) AddToMonitor(_ context.Context, _, _, _ string) error   { return nil }
func (s *stubTagRepo) RemoveFromMonitor(_ context.Context, _, _ string) error { return nil }

func TestTagProviderAdapter_ReturnsTagNames(t *testing.T) {
	repo := &stubTagRepo{
		monitorTags: []*tag.MonitorTag{
			{MonitorID: "mon-1", TagID: "t1"},
			{MonitorID: "mon-1", TagID: "t2"},
		},
		tags: map[string]*tag.Tag{
			"t1": {ID: "t1", Name: "production", Color: "#f50"},
			"t2": {ID: "t2", Name: "critical", Color: "#ff0000"},
		},
	}

	adapter := &tagProviderAdapter{tagRepo: repo}
	names, err := adapter.GetTagsForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	assert.Equal(t, []string{"production", "critical"}, names)
}

func TestTagProviderAdapter_SkipsMissingTag(t *testing.T) {
	repo := &stubTagRepo{
		monitorTags: []*tag.MonitorTag{
			{MonitorID: "mon-1", TagID: "t1"},
			{MonitorID: "mon-1", TagID: "t-missing"},
			{MonitorID: "mon-1", TagID: "t2"},
		},
		tags: map[string]*tag.Tag{
			"t1": {ID: "t1", Name: "env", Color: "#aaa"},
			"t2": {ID: "t2", Name: "region", Color: "#bbb"},
		},
	}

	adapter := &tagProviderAdapter{tagRepo: repo}
	names, err := adapter.GetTagsForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	assert.Equal(t, []string{"env", "region"}, names, "missing tag should be skipped silently")
}

func TestTagProviderAdapter_EmptyTags(t *testing.T) {
	repo := &stubTagRepo{
		monitorTags: nil,
		tags:        map[string]*tag.Tag{},
	}

	adapter := &tagProviderAdapter{tagRepo: repo}
	names, err := adapter.GetTagsForMonitor(t.Context(), "mon-1")
	require.NoError(t, err)
	assert.Empty(t, names)
}

// --- tagReaderAdapter ---

func TestTagReaderAdapter_ReturnsFullMonitorTagInfo(t *testing.T) {
	repo := &stubTagRepo{
		monitorTags: []*tag.MonitorTag{
			{MonitorID: "mon-1", TagID: "t1", Value: "us-east-1"},
			{MonitorID: "mon-1", TagID: "t2", Value: ""},
		},
		tags: map[string]*tag.Tag{
			"t1": {ID: "t1", Name: "region", Color: "#00ff00"},
			"t2": {ID: "t2", Name: "env", Color: "#0000ff"},
		},
	}

	adapter := &tagReaderAdapter{tagRepo: repo}
	infos, err := adapter.GetMonitorTags(t.Context(), "mon-1")
	require.NoError(t, err)
	require.Len(t, infos, 2)

	assert.Equal(t, domainmonitor.TagInfo{TagID: "t1", Name: "region", Color: "#00ff00", Value: "us-east-1"}, infos[0])
	assert.Equal(t, domainmonitor.TagInfo{TagID: "t2", Name: "env", Color: "#0000ff", Value: ""}, infos[1])
}

func TestTagReaderAdapter_SkipsMissingTag(t *testing.T) {
	repo := &stubTagRepo{
		monitorTags: []*tag.MonitorTag{
			{MonitorID: "mon-1", TagID: "t-gone"},
			{MonitorID: "mon-1", TagID: "t1"},
		},
		tags: map[string]*tag.Tag{
			"t1": {ID: "t1", Name: "valid", Color: "#ccc"},
		},
	}

	adapter := &tagReaderAdapter{tagRepo: repo}
	infos, err := adapter.GetMonitorTags(t.Context(), "mon-1")
	require.NoError(t, err)
	require.Len(t, infos, 1)
	assert.Equal(t, "valid", infos[0].Name)
}
