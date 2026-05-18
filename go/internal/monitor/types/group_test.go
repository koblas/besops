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

type mockTagFinder struct {
	monitors []*domainmonitor.Monitor
}

func (m *mockTagFinder) FindByTagIDs(_ context.Context, _ []string) ([]*domainmonitor.Monitor, error) {
	return m.monitors, nil
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
		TagFinder: &mockTagFinder{monitors: nil},
		HbStore:   &mockHbStore{beats: nil},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-1"}})
	require.NoError(t, err)
	assert.Equal(t, status.Pending, result.Status)
	assert.Equal(t, "Group empty", result.Message)
}

func TestGroupChecker_NoTagIDs(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "should-not-resolve", Name: "Ghost", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: nil},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group"})
	require.NoError(t, err)
	assert.Equal(t, status.Pending, result.Status)
	assert.Equal(t, "Group empty", result.Message)
}

func TestGroupChecker_AllChildrenUp(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "API", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Up)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-prod"}})
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
	assert.Equal(t, "All children up", result.Message)
}

func TestGroupChecker_OneChildDown(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "API", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Down)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-prod"}})
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
	assert.Contains(t, result.Message, "API")
}

func TestGroupChecker_InactiveChildIgnored(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "Paused", Active: false},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
			"child-2": {MonitorID: "child-2", Status: int(status.Down)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-a"}})
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
}

func TestGroupChecker_PendingChild(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "child-1", Name: "Web", Active: true},
				{ID: "child-2", Name: "New", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"child-1": {MonitorID: "child-1", Status: int(status.Up)},
		}},
	}

	result, err := checker.Check(t.Context(), &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-a"}})
	require.NoError(t, err)
	assert.Equal(t, status.Pending, result.Status)
	assert.Contains(t, result.Message, "New")
}

func TestGroupChecker_MultipleTagIDs(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "tagged-1", Name: "Tagged Web", Active: true},
				{ID: "tagged-2", Name: "Tagged API", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"tagged-1": {MonitorID: "tagged-1", Status: int(status.Up)},
			"tagged-2": {MonitorID: "tagged-2", Status: int(status.Up)},
		}},
	}

	cfg := &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-a", "tag-b"}}
	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
	assert.Equal(t, "All children up", result.Message)
}

func TestGroupChecker_TagBasedMixed(t *testing.T) {
	checker := &types.GroupChecker{
		TagFinder: &mockTagFinder{
			monitors: []*domainmonitor.Monitor{
				{ID: "tagged-1", Name: "Healthy", Active: true},
				{ID: "tagged-2", Name: "Broken", Active: true},
			},
		},
		HbStore: &mockHbStore{beats: map[string]*heartbeat.Heartbeat{
			"tagged-1": {MonitorID: "tagged-1", Status: int(status.Up)},
			"tagged-2": {MonitorID: "tagged-2", Status: int(status.Down)},
		}},
	}

	cfg := &monitor.Config{ID: "group-1", Type: "group", GroupTagIDs: []string{"tag-prod"}}
	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
	assert.Contains(t, result.Message, "Broken")
}
