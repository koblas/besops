package types

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

func TestTCPCheckerSuccess(t *testing.T) {
	// Start a local TCP listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}
			conn.Close()
		}
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	checker := &TCPChecker{}
	cfg := &monitor.Config{
		Hostname: "127.0.0.1",
		Port:     port,
		Timeout:  5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up, got %v: %s", result.Status, result.Message)
	}
	if result.Ping < 0 {
		t.Error("expected non-negative ping")
	}
}

func TestTCPCheckerRefused(t *testing.T) {
	checker := &TCPChecker{}
	cfg := &monitor.Config{
		Hostname: "127.0.0.1",
		Port:     1, // port 1 should be refused on localhost
		Timeout:  1 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down, got %v: %s", result.Status, result.Message)
	}
}
