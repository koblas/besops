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
	webhookURL, _ := config["discordWebhookUrl"].(string)
	if webhookURL == "" {
		return fmt.Errorf("discordWebhookUrl is required")
	}

	color := 0xe74c3c
	statusText := "Down"
	if heartbeat != nil && heartbeat.Status == 1 {
		color = 0x2ecc71
		statusText = "Up"
	}

	monitorName := ""
	if monitor != nil {
		monitorName = monitor.Name
	}

	embed := map[string]any{
		"title":       fmt.Sprintf("%s: %s", statusText, monitorName),
		"description": msg,
		"color":       color,
	}

	if heartbeat != nil && heartbeat.Time != "" {
		embed["timestamp"] = heartbeat.Time
	}

	payload := map[string]any{
		"username": "Bes Ops",
		"embeds":   []any{embed},
	}

	prefix, _ := config["discordPrefixMessage"].(string)
	if prefix != "" {
		payload["content"] = prefix
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
