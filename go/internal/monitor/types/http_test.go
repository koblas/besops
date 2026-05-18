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
	if result.Latency <= 0 {
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

func TestHTTPCheckerJsonPathMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"1.2.3"}`))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:           srv.URL,
		HTTP:          monitor.HTTPConfig{Method: "GET"},
		Timeout:       5 * time.Second,
		JsonPath:      "status",
		ExpectedValue: "ok",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (JSON path matches), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerJsonPathMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"error","code":500}`))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:           srv.URL,
		HTTP:          monitor.HTTPConfig{Method: "GET"},
		Timeout:       5 * time.Second,
		JsonPath:      "status",
		ExpectedValue: "ok",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down (JSON path mismatch), got %v: %s", result.Status, result.Message)
	}
	if !contains(result.Message, `expected "ok"`) {
		t.Errorf("expected error message to mention expected value, got: %s", result.Message)
	}
}

func TestHTTPCheckerJsonPathNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:      srv.URL,
		HTTP:     monitor.HTTPConfig{Method: "GET"},
		Timeout:  5 * time.Second,
		JsonPath: "status",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down (JSON path not found), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerJsonPathExistsWithoutExpected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alive":true}`))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:      srv.URL,
		HTTP:     monitor.HTTPConfig{Method: "GET"},
		Timeout:  5 * time.Second,
		JsonPath: "alive",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (JSON path exists, no expected value), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerJsonPathNestedObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"db":{"connected":true},"version":"2.0"}}`))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:           srv.URL,
		HTTP:          monitor.HTTPConfig{Method: "GET"},
		Timeout:       5 * time.Second,
		JsonPath:      "data.db.connected",
		ExpectedValue: "true",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Up {
		t.Errorf("expected Up (nested JSON path matches), got %v: %s", result.Status, result.Message)
	}
}

func TestHTTPCheckerJsonPathInvalidJson(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json at all"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:      srv.URL,
		HTTP:     monitor.HTTPConfig{Method: "GET"},
		Timeout:  5 * time.Second,
		JsonPath: "status",
	}

	result, err := checker.Check(t.Context(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != status.Down {
		t.Errorf("expected Down (invalid JSON), got %v: %s", result.Status, result.Message)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
