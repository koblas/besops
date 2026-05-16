package types

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"tailscale.com/client/local"
	"tailscale.com/tailcfg"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type TailscalePingChecker struct{}

func (c *TailscalePingChecker) Type() string { return "tailscale-ping" }

func (c *TailscalePingChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	addr, err := netip.ParseAddr(cfg.Hostname)
	if err != nil {
		// Try resolving as hostname via Tailscale's MagicDNS — not directly supported,
		// so we require an IP address
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("invalid Tailscale IP address: %v", err),
		}, nil
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	lc := &local.Client{}

	start := time.Now()
	result, err := lc.Ping(ctx, addr, tailcfg.PingDisco)
	ping := time.Since(start).Milliseconds()

	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("tailscale ping failed: %v", err),
		}, nil
	}

	if result.Err != "" {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("tailscale ping error: %s", result.Err),
		}, nil
	}

	latencyMs := int64(result.LatencySeconds * 1000)

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    latencyMs,
		Message: fmt.Sprintf("pong from %s via %s", result.NodeName, endpoint(result.Endpoint)),
	}, nil
}

func endpoint(ep string) string {
	if ep == "" {
		return "relay"
	}
	return ep
}
