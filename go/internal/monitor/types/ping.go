package types

import (
	"context"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type PingChecker struct{}

func (c *PingChecker) Type() string { return "ping" }

func (c *PingChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	count := 3
	if cfg.Port > 0 && cfg.Port <= 10 {
		count = cfg.Port // overloaded: packet count via Port field
	}

	pinger, err := probing.NewPinger(cfg.Hostname)
	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("ping setup failed: %v", err),
		}, nil
	}

	pinger.Count = count
	pinger.Timeout = timeout
	pinger.SetPrivileged(false)

	done := make(chan error, 1)
	go func() {
		done <- pinger.Run()
	}()

	select {
	case <-ctx.Done():
		pinger.Stop()
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "ping cancelled",
		}, nil
	case runErr := <-done:
		if runErr != nil {
			return monitor.CheckResult{
				Status:  status.Down,
				Message: fmt.Sprintf("ping failed: %v", runErr),
			}, nil
		}
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    stats.AvgRtt.Milliseconds(),
			Message: fmt.Sprintf("0/%d packets received", stats.PacketsSent),
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    stats.AvgRtt.Milliseconds(),
		Message: fmt.Sprintf("%d/%d packets, avg %.1fms", stats.PacketsRecv, stats.PacketsSent, float64(stats.AvgRtt.Microseconds())/1000),
	}, nil
}
