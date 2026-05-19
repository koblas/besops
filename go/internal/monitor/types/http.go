package types

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
	"github.com/tidwall/gjson"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker() *HTTPChecker {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &HTTPChecker{
		client: &http.Client{
			Transport: transport,
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (c *HTTPChecker) Type() string { return "http" }

func (c *HTTPChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Per-request TLS config via a cloned transport
	transport := c.client.Transport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = cfg.IgnoreTLS //nolint:gosec // user-configured
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: c.client.CheckRedirect,
	}

	method := cfg.HTTP.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if cfg.HTTP.Body != "" {
		bodyReader = strings.NewReader(cfg.HTTP.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, cfg.URL, bodyReader)
	if err != nil {
		return monitor.CheckResult{Status: status.Down, Message: fmt.Sprintf("invalid request: %v", err)}, nil
	}

	for _, h := range cfg.HTTP.Headers {
		req.Header.Add(h.Name, h.Value)
		slog.DebugContext(ctx, "setting header", slog.String("key", h.Name), slog.String("value", h.Value))
	}
	if cfg.HTTP.BasicAuthUser != "" {
		req.SetBasicAuth(cfg.HTTP.BasicAuthUser, cfg.HTTP.BasicAuthPass)
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return monitor.CheckResult{Status: status.Down, Latency: latency, Message: err.Error()}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max

	result := monitor.CheckResult{
		Latency:      latency,
		ResponseBody: body,
		Message:      fmt.Sprintf("%d - %s", resp.StatusCode, string(body)[:min(len(body), 80)]),
	}

	// TLS certificate info
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
		result.CertInfo = &monitor.CertInfo{
			Valid:    time.Now().Before(cert.NotAfter) && time.Now().After(cert.NotBefore),
			Issuer:   cert.Issuer.CommonName,
			Subject:  cert.Subject.CommonName,
			DaysLeft: daysLeft,
			ExpiryAt: cert.NotAfter.Format(time.RFC3339),
		}
	}

	// Status code check
	if !isAcceptedStatus(resp.StatusCode, cfg.HTTP.AcceptedStatusCodes) {
		result.Status = status.Down
		return result, nil
	}

	// Keyword check
	if cfg.Keyword != "" {
		bodyStr := string(body)
		found := strings.Contains(bodyStr, cfg.Keyword)
		if cfg.KeywordType == "not contain" {
			found = !found
		}
		if !found {
			result.Status = status.Down
			if cfg.KeywordType == "not contain" {
				result.Message = fmt.Sprintf("keyword %q found (should not contain)", cfg.Keyword)
			} else {
				result.Message = fmt.Sprintf("keyword %q not found in response", cfg.Keyword)
			}
			return result, nil
		}
	}

	// JSON path check
	if cfg.JsonPath != "" {
		bodyStr := string(body)
		if !gjson.Valid(bodyStr) {
			result.Status = status.Down
			result.Message = "response body is not valid JSON"
			return result, nil
		}

		got := gjson.Get(bodyStr, cfg.JsonPath)
		if !got.Exists() {
			result.Status = status.Down
			result.Message = fmt.Sprintf("JSON path %q not found in response", cfg.JsonPath)
			return result, nil
		}

		if cfg.ExpectedValue != "" && got.String() != cfg.ExpectedValue {
			result.Status = status.Down
			result.Message = fmt.Sprintf("JSON path %q returned %q, expected %q", cfg.JsonPath, got.String(), cfg.ExpectedValue)
			return result, nil
		}
	}

	result.Status = status.Up
	return result, nil
}

func isAcceptedStatus(code int, accepted []string) bool {
	if len(accepted) == 0 {
		return code >= 200 && code < 400
	}
	codeStr := strconv.Itoa(code)
	for _, s := range accepted {
		s = strings.TrimSpace(s)
		if s == codeStr {
			return true
		}
		// Range like "200-299"
		if parts := strings.SplitN(s, "-", 2); len(parts) == 2 {
			lo, loErr := strconv.Atoi(parts[0])
			hi, hiErr := strconv.Atoi(parts[1])
			if loErr == nil && hiErr == nil && code >= lo && code <= hi {
				return true
			}
		}
		// Pattern like "2xx" matches 200-299
		if len(s) == 3 && strings.HasSuffix(s, "xx") {
			if codeStr[0] == s[0] {
				return true
			}
		}
	}
	return false
}
