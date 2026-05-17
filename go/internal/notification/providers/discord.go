package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/koblas/besops/internal/notification"
)

type DiscordNotifier struct{}

func (n *DiscordNotifier) Name() string { return "discord" }

func (n *DiscordNotifier) Send(ctx context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, heartbeat *notification.HeartbeatInfo) error {
	webhookURL, _ := config["discordWebhookURL"].(string)
	if webhookURL == "" {
		return fmt.Errorf("discordWebhookURL is required")
	}

	color := 0xe74c3c // red
	if heartbeat != nil && heartbeat.Status == 1 {
		color = 0x2ecc71 // green
	}

	username, _ := config["discordUsername"].(string)
	if username == "" {
		username = "Bes Ops"
	}

	monitorName := ""
	monitorURL := ""
	if monitor != nil {
		monitorName = monitor.Name
		monitorURL = monitor.URL
	}

	embed := map[string]any{
		"title":       monitorName,
		"description": msg,
		"color":       color,
	}
	if monitorURL != "" {
		embed["url"] = monitorURL
	}

	payload := map[string]any{
		"username": username,
		"embeds":   []any{embed},
	}

	prefixContent, _ := config["discordPrefixMessage"].(string)
	if prefixContent != "" {
		payload["content"] = prefixContent
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}

	return nil
}
