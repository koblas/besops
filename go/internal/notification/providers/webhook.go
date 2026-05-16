package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/koblas/besops/internal/notification"
)

type WebhookNotifier struct{}

func (n *WebhookNotifier) Name() string { return "webhook" }

func (n *WebhookNotifier) Send(ctx context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, heartbeat *notification.HeartbeatInfo) error {
	url, _ := config["webhookURL"].(string)
	if url == "" {
		return fmt.Errorf("webhookURL is required")
	}

	contentType, _ := config["webhookContentType"].(string)
	if contentType == "" {
		contentType = "application/json"
	}

	payload := map[string]any{
		"msg":       msg,
		"heartbeat": heartbeat,
		"monitor":   monitor,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	if headers, ok := config["webhookAdditionalHeaders"].(string); ok && headers != "" {
		var extra map[string]string
		if jsonErr := json.Unmarshal([]byte(headers), &extra); jsonErr == nil {
			for k, v := range extra {
				req.Header.Set(k, v)
			}
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
