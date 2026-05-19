package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

func TestDNSCheckerA(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "google.com",
		DNS:     monitor.DNSConfig{ResolveType: "A"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
	assert.NotEmpty(t, result.Message, "expected non-empty message with IP addresses")
}

func TestDNSCheckerMX(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "google.com",
		DNS:     monitor.DNSConfig{ResolveType: "MX"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
}

func TestDNSCheckerNonexistent(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "this-domain-does-not-exist-xyzzy.example",
		DNS:     monitor.DNSConfig{ResolveType: "A"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
}

func TestDNSCheckerUnsupportedType(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "google.com",
		DNS:     monitor.DNSConfig{ResolveType: "INVALID"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
}
