package types

import (
	"context"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

// PushChecker is a passive monitor — it doesn't actively probe anything.
// The push endpoint handles incoming heartbeats. When the monitor loop calls
// Check, it simply reports UP since the loop itself tracks staleness via
// missed intervals.
type PushChecker struct{}

func (c *PushChecker) Type() string { return "push" }

func (c *PushChecker) Check(_ context.Context, _ *monitor.Config) (monitor.CheckResult, error) {
	return monitor.CheckResult{
		Status:  status.Up,
		Message: "waiting for push",
	}, nil
}
