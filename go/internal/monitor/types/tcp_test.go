package types

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

func TestTCPCheckerSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
	assert.GreaterOrEqual(t, result.Latency, int64(0))
}

func TestTCPCheckerRefused(t *testing.T) {
	checker := &TCPChecker{}
	cfg := &monitor.Config{
		Hostname: "127.0.0.1",
		Port:     1,
		Timeout:  1 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status, result.Message)
}
