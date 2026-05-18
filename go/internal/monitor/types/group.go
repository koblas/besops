package types

import (
	"context"
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
func (c *GroupChecker) IsAggregate()  {}

func (c *GroupChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	members, err := c.resolveMembers(ctx, cfg)
	if err != nil {
		return monitor.CheckResult{Status: status.Down, Message: "failed to load members"}, nil
	}

	if len(members) == 0 {
		return monitor.CheckResult{Status: status.Pending, Message: "Group empty"}, nil
	}

	worstStatus := status.Up
	var downNames []string
	var pendingNames []string

	for _, child := range members {
		if !child.Active {
			continue
		}

		latest, hbErr := c.HbStore.GetLatest(ctx, child.ID)
		if hbErr != nil || latest == nil {
			if worstStatus == status.Up {
				worstStatus = status.Pending
			}
			pendingNames = append(pendingNames, child.Name)
			continue
		}

		childStatus := status.Status(latest.Status)
		if childStatus == status.Down {
			worstStatus = status.Down
			downNames = append(downNames, child.Name)
		} else if childStatus == status.Pending {
			if worstStatus != status.Down {
				worstStatus = status.Pending
			}
			pendingNames = append(pendingNames, child.Name)
		}
	}

	switch worstStatus {
	case status.Up:
		return monitor.CheckResult{Status: status.Up, Message: "All children up"}, nil
	case status.Pending:
		return monitor.CheckResult{Status: status.Pending, Message: "Pending: " + strings.Join(pendingNames, ", ")}, nil
	default:
		msg := "Down: " + strings.Join(downNames, ", ")
		if len(pendingNames) > 0 {
			msg += "; pending: " + strings.Join(pendingNames, ", ")
		}
		return monitor.CheckResult{Status: status.Down, Message: msg}, nil
	}
}

func (c *GroupChecker) resolveMembers(ctx context.Context, cfg *monitor.Config) ([]*domainmonitor.Monitor, error) {
	if len(cfg.GroupTagIDs) > 0 && c.TagFinder != nil {
		return c.TagFinder.FindByTagIDs(ctx, cfg.GroupTagIDs)
	}
	return nil, nil
}
