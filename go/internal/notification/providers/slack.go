package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/koblas/besops/internal/notification"
)

type SlackNotifier struct{}

func (n *SlackNotifier) Name() string { return "slack" }

func (n *SlackNotifier) Send(ctx context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, heartbeat *notification.HeartbeatInfo) error {
	webhookURL, _ := config["slackwebhookURL"].(string)
	if webhookURL == "" {
		return fmt.Errorf("slackwebhookURL is required")
	}

	color := "#e74c3c"
	statusText := "Down"
	if heartbeat != nil && heartbeat.Status == 1 {
		color = "#2ecc71"
		statusText = "Up"
	}

	channel, _ := config["slackchannel"].(string)
	username, _ := config["slackusername"].(string)
	if username == "" {
		username = "Bes Ops"
	}
	iconEmoji, _ := config["slackiconemo"].(string)

	monitorName := ""
	monitorURL := ""
	if monitor != nil {
		monitorName = monitor.Name
		monitorURL = monitor.URL
	}

	attachment := map[string]any{
		"color":  color,
		"title":  fmt.Sprintf("%s: %s", statusText, monitorName),
		"text":   msg,
		"fields": []map[string]any{},
	}
	if monitorURL != "" {
		attachment["title_link"] = monitorURL
	}

	payload := map[string]any{
		"username":    username,
		"attachments": []any{attachment},
	}
	if channel != "" {
		payload["channel"] = channel
	}
	if iconEmoji != "" {
		payload["icon_emoji"] = iconEmoji
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}
