package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/koblas/besops/internal/notification"
)

type TelegramNotifier struct{}

func (n *TelegramNotifier) Name() string { return "telegram" }

func (n *TelegramNotifier) Send(ctx context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, heartbeat *notification.HeartbeatInfo) error {
	botToken, _ := config["telegramBotToken"].(string)
	if botToken == "" {
		return fmt.Errorf("telegramBotToken is required")
	}

	chatID, _ := config["telegramChatID"].(string)
	if chatID == "" {
		return fmt.Errorf("telegramChatID is required")
	}

	threadID, _ := config["telegramMessageThreadID"].(string)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	}
	if threadID != "" {
		payload["message_thread_id"] = threadID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}
