package types

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

func TestHTTPCheckerSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET"},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up, got %v: %s", result.Status, result.Message)
	}
	if result.Ping <= 0 {
		t.Error("expected positive ping")
	}
}

func TestHTTPCheckerKeywordContain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("the service is healthy"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:         srv.URL,
		HTTP:        monitor.HTTPConfig{Method: "GET"},
		Timeout:     5 * time.Second,
		Keyword:     "healthy",
		KeywordType: "contain",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (keyword found), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerKeywordNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("error occurred"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:         srv.URL,
		HTTP:        monitor.HTTPConfig{Method: "GET"},
		Timeout:     5 * time.Second,
		Keyword:     "healthy",
		KeywordType: "contain",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down (keyword not found), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerKeywordNotContain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("all good"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:         srv.URL,
		HTTP:        monitor.HTTPConfig{Method: "GET"},
		Timeout:     5 * time.Second,
		Keyword:     "error",
		KeywordType: "not contain",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (keyword absent), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerBadStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET", AcceptedStatusCodes: []string{"200-299"}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down for 500, got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerCustomAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET", AcceptedStatusCodes: []string{"404"}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (404 accepted), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerRangeAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET", AcceptedStatusCodes: []string{"200-299"}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (201 in 200-299 range), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerHeaders(t *testing.T) {
	var receivedHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET", Headers: map[string]string{"X-Custom": "test-value"}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up, got %v", result.Status)
	}
	if receivedHeader != "test-value" {
		t.Errorf("expected header test-value, got %s", receivedHeader)
	}
}

func TestHTTPCheckerTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET"},
		Timeout: 100 * time.Millisecond,
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down on timeout, got %v: %s", result.Status, result.Message)
	}
}
