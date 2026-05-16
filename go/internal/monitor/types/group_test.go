package types_test

import (
	"context"
	"testing"

	"github.com/koblas/besops/internal/domain/heartbeat"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/internal/monitor/types"
	"github.com/koblas/besops/lib/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMonitorRepo struct {
	children []*domainmonitor.Monitor
}

func (m *mockMonitorRepo) FindByID(_ context.Context, _ string) (*domainmonitor.Monitor, error) {
	return nil, nil
}
func (m *mockMonitorRepo) FindByUserID(_ context.Context, _ string) ([]*domainmonitor.Monitor, error) {
	return nil, nil
}
func (m *mockMonitorRepo) FindByPushToken(_ context.Context, _ string) (*domainmonitor.Monitor, error) {
	return nil, nil
}
func (m *mockMonitorRepo) FindAllActiveIDs(_ context.Context) ([]string, error) { return nil, nil }
func (m *mockMonitorRepo) Create(_ context.Context, _ *domainmonitor.Monitor) (string, error) {
	return "", nil
}
func (m *mockMonitorRepo) Update(_ context.Context, _ *domainmonitor.Monitor) error { return nil }
func (m *mockMonitorRepo) Delete(_ context.Context, _ string) error                 { return nil }
func (m *mockMonitorRepo) GetChildren(_ context.Context, _ string) ([]*domainmonitor.Monitor, error) {
	return m.children, nil
}

type mockHbStore struct {
	beats map[string]*heartbeat.Heartbeat
}

func (m *mockHbStore) GetLatest(_ context.Context, monitorID string) (*heartbeat.Heartbeat, error) {
	hb, ok := m.beats[monitorID]
	if !ok {
		return nil, nil
	}
	return hb, nil
}

func TestGroupChecker_EmptyGroup(t *testing.T) {
	checker := &types.GroupChecker{
		MonitorRepo: &mockMonitorRepo{children: nil},
		HbStore:     &mockHbStore{beats: nil},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Pending, result.Status)
	assert.Equal(t, "Group empty", result.Message)
}

func TestGroupChecker_AllChildrenUp(t *testing.T) {
	checker := &types.GroupChecker{
		MonitorRepo: &mockMonitorRepo{
			children: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "API", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Up)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
	assert.Equal(t, "All children up", result.Message)
}

func TestGroupChecker_OneChildDown(t *testing.T) {
	checker := &types.GroupChecker{
		MonitorRepo: &mockMonitorRepo{
			children: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "API", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Down)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
	assert.Contains(t, result.Message, "API")
}

func TestGroupChecker_InactiveChildIgnored(t *testing.T) {
	checker := &types.GroupChecker{
		MonitorRepo: &mockMonitorRepo{
			children: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "Paused", Active: false},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Down)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
}

func TestGroupChecker_PendingChild(t *testing.T) {
	checker := &types.GroupChecker{
		MonitorRepo: &mockMonitorRepo{
			children: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "New", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Pending, result.Status)
	assert.Contains(t, result.Message, "New")
}
