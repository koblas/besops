package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/koblas/besops/internal/notification"
)

type PagerDutyNotifier struct{}

func (n *PagerDutyNotifier) Name() string { return "pagerduty" }

func (n *PagerDutyNotifier) Send(ctx context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, heartbeat *notification.HeartbeatInfo) error {
	integrationKey, _ := config["pagerdutyIntegrationKey"].(string)
	if integrationKey == "" {
		return fmt.Errorf("pagerdutyIntegrationKey is required")
	}

	severity, _ := config["pagerdutySeverity"].(string)
	if severity == "" {
		severity = "critical"
	}

	eventAction := "trigger"
	if heartbeat != nil && heartbeat.Status == 1 {
		eventAction = "resolve"
	}

	source := "besops"
	summary := msg
	if monitor != nil {
		source = monitor.Name
	}

	dedupKey := ""
	if monitor != nil {
		dedupKey = fmt.Sprintf("besops-%s", monitor.Name)
	}

	payload := map[string]any{
		"routing_key":  integrationKey,
		"event_action": eventAction,
		"dedup_key":    dedupKey,
		"payload": map[string]any{
			"summary":  summary,
			"source":   source,
			"severity": severity,
		},
	}

	if monitor != nil && monitor.URL != "" {
		payload["client"] = "Bes Ops"
		payload["client_url"] = monitor.URL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling pagerduty payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://events.pagerduty.com/v2/enqueue", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating pagerduty request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending pagerduty event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pagerduty returned status %d", resp.StatusCode)
	}

	return nil
}
