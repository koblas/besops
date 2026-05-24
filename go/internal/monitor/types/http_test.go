package types

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
	assert.Positive(t, result.Latency)
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
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
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
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status, result.Message)
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
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
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
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status, result.Message)
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
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
}

func TestHTTPCheckerRangeAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:     srv.URL,
		HTTP:    monitor.HTTPConfig{Method: "GET", AcceptedStatusCodes: []string{"200-299"}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
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
		HTTP:    monitor.HTTPConfig{Method: "GET", Headers: []monitor.HeaderPair{{Name: "X-Custom", Value: "test-value"}}},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
	assert.Equal(t, "test-value", receivedHeader)
}

func TestHTTPCheckerDuplicateHeaders(t *testing.T) {
	var receivedValues []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedValues = r.Header.Values("X-Multi")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL: srv.URL,
		HTTP: monitor.HTTPConfig{
			Method: "GET",
			Headers: []monitor.HeaderPair{
				{Name: "X-Multi", Value: "first"},
				{Name: "X-Multi", Value: "second"},
			},
		},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status)
	assert.Equal(t, []string{"first", "second"}, receivedValues)
}

func TestHTTPCheckerJSONPathMatch(t *testing.T) {
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
		JSONPath:      "status",
		ExpectedValue: "ok",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
}

func TestHTTPCheckerJSONPathMismatch(t *testing.T) {
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
		JSONPath:      "status",
		ExpectedValue: "ok",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
	assert.Contains(t, result.Message, `expected "ok"`)
}

func TestHTTPCheckerJSONPathNotFound(t *testing.T) {
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
		JSONPath: "status",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
}

func TestHTTPCheckerJSONPathExistsWithoutExpected(t *testing.T) {
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
		JSONPath: "alive",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
}

func TestHTTPCheckerJSONPathNestedObject(t *testing.T) {
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
		JSONPath:      "data.db.connected",
		ExpectedValue: "true",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
}

func TestHTTPCheckerJSONPathInvalidJson(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json at all"))
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL:      srv.URL,
		HTTP:     monitor.HTTPConfig{Method: "GET"},
		Timeout:  5 * time.Second,
		JSONPath: "status",
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status)
}

func TestHTTPCheckerExplicitContentType(t *testing.T) {
	var receivedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	checker := NewHTTPChecker()
	cfg := &monitor.Config{
		URL: srv.URL,
		HTTP: monitor.HTTPConfig{
			Method:  "POST",
			Body:    "name=value",
			Headers: []monitor.HeaderPair{{Name: "Content-Type", Value: "application/x-www-form-urlencoded"}},
		},
		Timeout: 5 * time.Second,
	}

	result, err := checker.Check(t.Context(), cfg)
	require.NoError(t, err)
	assert.Equal(t, status.Up, result.Status, result.Message)
	assert.Equal(t, "application/x-www-form-urlencoded", receivedContentType)
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
	require.NoError(t, err)
	assert.Equal(t, status.Down, result.Status, result.Message)
}
