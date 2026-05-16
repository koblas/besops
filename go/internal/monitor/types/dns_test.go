package types

import (
	"testing"
	"time"

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up, got %v: %s", result.Status, result.Message)
	}
	if result.Message == "" {
		t.Error("expected non-empty message with IP addresses")
	}
}

func TestDNSCheckerMX(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "google.com",
		DNS:     monitor.DNSConfig{ResolveType: "MX"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up, got %v: %s", result.Status, result.Message)
	}
}

func TestDNSCheckerNonexistent(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "this-domain-does-not-exist-xyzzy.example",
		DNS:     monitor.DNSConfig{ResolveType: "A"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down for nonexistent domain, got %v", result.Status)
	}
}

func TestDNSCheckerUnsupportedType(t *testing.T) {
	checker := &DNSChecker{}
	cfg := &monitor.Config{
		URL:     "google.com",
		DNS:     monitor.DNSConfig{ResolveType: "INVALID"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down for unsupported type, got %v", result.Status)
	}
}
