package types

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/koblas/besops/internal/domain/heartbeat"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

// GroupHeartbeatReader is the subset of HeartbeatStore needed by the group checker.
type GroupHeartbeatReader interface {
	GetLatest(ctx context.Context, monitorID string) (*heartbeat.Heartbeat, error)
}

// TagMonitorFinder resolves tag IDs to the monitors that carry them.
type TagMonitorFinder interface {
	FindByTagIDs(ctx context.Context, tagIDs []string) ([]*domainmonitor.Monitor, error)
}

// GroupChecker aggregates the status of member monitors to produce a group heartbeat.
// Members are resolved by tags via TagFinder.
type GroupChecker struct {
	HbStore   GroupHeartbeatReader
	TagFinder TagMonitorFinder
}

func (c *GroupChecker) Type() string { return "group" }
func (c *GroupChecker) IsAggregate() {}

func (c *GroupChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	slog.DebugContext(ctx, "group check starting", slog.String("monitor", cfg.ID), slog.Int("tag_count", len(cfg.GroupTagIDs)))

	if len(cfg.GroupTagIDs) == 0 {
		return monitor.CheckResult{Status: status.Pending, Message: "No member tags configured"}, nil
	}

	members, err := c.resolveMembers(ctx, cfg)
	if err != nil {
		slog.WarnContext(ctx, "group member resolution failed", slog.String("monitor", cfg.ID), slog.Any("error", err))
		return monitor.CheckResult{Status: status.Down, Message: "failed to load members"}, nil
	}

	if len(members) == 0 {
		slog.DebugContext(ctx, "group has no members", slog.String("monitor", cfg.ID), slog.Any("tag_ids", cfg.GroupTagIDs))
		return monitor.CheckResult{Status: status.Pending, Message: "No monitors have the selected tags"}, nil
	}

	var upCount, downCount int
	var downNames []string
	var pendingNames []string

	for _, child := range members {
		if !child.Active {
			continue
		}

		latest, hbErr := c.HbStore.GetLatest(ctx, child.ID)
		if hbErr != nil || latest == nil {
			pendingNames = append(pendingNames, child.Name)
			continue
		}

		childStatus := status.Status(latest.Status)
		switch childStatus {
		case status.Down:
			downCount++
			downNames = append(downNames, child.Name)
		case status.Up:
			upCount++
		default:
			pendingNames = append(pendingNames, child.Name)
		}
	}

	activeCount := upCount + downCount + len(pendingNames)

	if activeCount == 0 {
		return monitor.CheckResult{Status: status.Pending, Message: "No active children"}, nil
	}

	if downCount == 0 && len(pendingNames) == 0 {
		return monitor.CheckResult{Status: status.Up, Message: "All children up"}, nil
	}

	if downCount == 0 {
		if upCount > 0 {
			return monitor.CheckResult{Status: status.Up, Message: "All children up"}, nil
		}
		return monitor.CheckResult{Status: status.Pending, Message: "Pending: " + strings.Join(pendingNames, ", ")}, nil
	}

	if upCount == 0 && len(pendingNames) == 0 {
		return monitor.CheckResult{Status: status.Down, Message: "Down: " + strings.Join(downNames, ", ")}, nil
	}

	// Some children are down but others are up or pending — degraded
	msg := "Down: " + strings.Join(downNames, ", ")
	if len(pendingNames) > 0 {
		msg += "; pending: " + strings.Join(pendingNames, ", ")
	}
	return monitor.CheckResult{Status: status.Degraded, Message: msg}, nil
}

func (c *GroupChecker) resolveMembers(ctx context.Context, cfg *monitor.Config) ([]*domainmonitor.Monitor, error) {
	members, err := c.TagFinder.FindByTagIDs(ctx, cfg.GroupTagIDs)
	if err != nil {
		return nil, fmt.Errorf("resolving group members by tag IDs: %w", err)
	}
	return members, nil
}
