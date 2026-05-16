package types

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type SMTPMonitorChecker struct{}

func (c *SMTPMonitorChecker) Type() string { return "smtp" }

func (c *SMTPMonitorChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	port := cfg.Port
	if port == 0 {
		port = 25
	}

	addr := fmt.Sprintf("%s:%d", cfg.Hostname, port)

	start := time.Now()

	var client *smtp.Client
	var connErr error

	if cfg.SMTP.Security == "secure" {
		tlsConfig := &tls.Config{
			ServerName:         cfg.Hostname,
			InsecureSkipVerify: cfg.IgnoreTLS, //nolint:gosec // user-configurable
		}
		dialer := &tls.Dialer{Config: tlsConfig, NetDialer: &net.Dialer{Timeout: timeout}}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			ping := time.Since(start).Milliseconds()
			return monitor.CheckResult{
				Status:  status.Down,
				Ping:    ping,
				Message: fmt.Sprintf("TLS connection failed: %v", err),
			}, nil
		}
		client, connErr = smtp.NewClient(conn, cfg.Hostname)
		if connErr != nil {
			conn.Close()
		}
	} else {
		dialer := &net.Dialer{Timeout: timeout}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			ping := time.Since(start).Milliseconds()
			return monitor.CheckResult{
				Status:  status.Down,
				Ping:    ping,
				Message: fmt.Sprintf("connection failed: %v", err),
			}, nil
		}
		client, connErr = smtp.NewClient(conn, cfg.Hostname)
		if connErr != nil {
			conn.Close()
		}
	}

	ping := time.Since(start).Milliseconds()

	if connErr != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("SMTP client creation failed: %v", connErr),
		}, nil
	}
	defer client.Close()

	if cfg.SMTP.Security == "starttls" {
		tlsConfig := &tls.Config{
			ServerName:         cfg.Hostname,
			InsecureSkipVerify: cfg.IgnoreTLS, //nolint:gosec // user-configurable
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return monitor.CheckResult{
				Status:  status.Down,
				Ping:    ping,
				Message: fmt.Sprintf("STARTTLS failed: %v", err),
			}, nil
		}
	}

	if err := client.Noop(); err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("SMTP NOOP failed: %v", err),
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    ping,
		Message: "SMTP connection verifies successfully",
	}, nil
}
