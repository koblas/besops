package types

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type TCPChecker struct{}

func (c *TCPChecker) Type() string { return "port" }

func (c *TCPChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	addr := fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)

	start := time.Now()
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", addr)
	ping := time.Since(start).Milliseconds()

	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("connection failed: %v", err),
		}, nil
	}
	conn.Close()

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    ping,
		Message: fmt.Sprintf("TCP connection to %s successful", addr),
	}, nil
}
