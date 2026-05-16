package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/koblas/besops/internal/auth"
	"github.com/koblas/besops/internal/broadcast"
)

type wsEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func NewWebSocketHandler(events broadcast.Subscriber, authProvider *auth.Provider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractWSToken(r)
		if token == "" {
			slog.DebugContext(r.Context(), "ws: missing token in subprotocol header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := authProvider.ValidateAccessToken(token); err != nil {
			slog.DebugContext(r.Context(), "ws: invalid token", slog.Any("error", err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			Subprotocols:       []string{"bearer"},
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "ws: accept failed", slog.Any("error", err))
			return
		}
		defer conn.CloseNow()

		ctx := r.Context()
		sub := events.Subscribe()
		defer events.Unsubscribe(sub)

		slog.DebugContext(ctx, "ws: client connected")

		for {
			select {
			case <-ctx.Done():
				slog.DebugContext(ctx, "ws: context cancelled, closing")
				conn.Close(websocket.StatusNormalClosure, "")
				return
			case ev, ok := <-sub:
				if !ok {
					slog.DebugContext(ctx, "ws: event channel closed")
					conn.Close(websocket.StatusNormalClosure, "")
					return
				}
				msg, marshalErr := json.Marshal(wsEvent{Type: ev.Type, Data: ev.Data})
				if marshalErr != nil {
					slog.DebugContext(ctx, "ws: marshal failed", slog.String("type", ev.Type), slog.Any("error", marshalErr))
					continue
				}
				if writeErr := conn.Write(ctx, websocket.MessageText, msg); writeErr != nil {
					slog.DebugContext(ctx, "ws: write failed, disconnecting", slog.Any("error", writeErr))
					return
				}
			}
		}
	})
}

// extractWSToken gets the JWT from the WebSocket subprotocol header.
// The client sends: Sec-WebSocket-Protocol: bearer, <token>
func extractWSToken(r *http.Request) string {
	protocols := r.Header.Values("Sec-WebSocket-Protocol")
	for _, p := range protocols {
		for _, sub := range strings.Split(p, ",") {
			sub = strings.TrimSpace(sub)
			if sub != "bearer" && sub != "" {
				return sub
			}
		}
	}
	return ""
}
