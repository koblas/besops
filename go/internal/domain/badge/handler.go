package badge

import (
	"context"
	"fmt"
	"strings"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

type HeartbeatQuerier interface {
	GetAverageLatency(ctx context.Context, monitorID string, hours int) (float64, error)
	GetAverageResponse(ctx context.Context, monitorID string, hours int) (float64, error)
	GetUptime(ctx context.Context, monitorID string, hours int) (float64, error)
}

var _ oas.BadgeHandler = (*Handler)(nil)

type Handler struct {
	heartbeats HeartbeatQuerier
}

func NewHandler(hq HeartbeatQuerier) *Handler {
	return &Handler{heartbeats: hq}
}

func (h *Handler) GetStatusBadge(ctx context.Context, params oas.GetStatusBadgeParams) (oas.SVGBadge, error) {
	uptime, err := h.heartbeats.GetUptime(ctx, params.MonitorId.String(), 24)
	if err != nil {
		return oas.SVGBadge{Data: strings.NewReader(renderBadge("Status", "Unknown", "#999"))}, nil //nolint:nilerr
	}

	label := "Up"
	color := "#4c1"
	if uptime < 1.0 {
		label = "Down"
		color = "#e05d44"
	}

	return oas.SVGBadge{Data: strings.NewReader(renderBadge("Status", label, color))}, nil
}

func (h *Handler) GetUptimeBadge(ctx context.Context, params oas.GetUptimeBadgeParams) (oas.SVGBadge, error) {
	hours := 24
	if params.Duration.IsSet() {
		hours = int(params.Duration.Value)
	}

	uptime, err := h.heartbeats.GetUptime(ctx, params.MonitorId.String(), hours)
	if err != nil {
		return oas.SVGBadge{Data: strings.NewReader(renderBadge("Uptime", "N/A", "#999"))}, nil //nolint:nilerr
	}

	value := fmt.Sprintf("%.1f%%", uptime*100)
	color := "#4c1"
	if uptime < 0.99 {
		color = "#dfb317"
	}
	if uptime < 0.95 {
		color = "#e05d44"
	}

	return oas.SVGBadge{Data: strings.NewReader(renderBadge("Uptime", value, color))}, nil
}

func (h *Handler) GetLatencyBadge(ctx context.Context, params oas.GetLatencyBadgeParams) (oas.SVGBadge, error) {
	hours := 24
	if params.Duration.IsSet() {
		hours = int(params.Duration.Value)
	}

	latency, err := h.heartbeats.GetAverageLatency(ctx, params.MonitorId.String(), hours)
	if err != nil {
		return oas.SVGBadge{Data: strings.NewReader(renderBadge("Latency", "N/A", "#999"))}, nil //nolint:nilerr
	}

	value := fmt.Sprintf("%.0fms", latency)
	return oas.SVGBadge{Data: strings.NewReader(renderBadge("Latency", value, "#4c1"))}, nil
}

func (h *Handler) GetResponseBadge(ctx context.Context, params oas.GetResponseBadgeParams) (oas.SVGBadge, error) {
	resp, err := h.heartbeats.GetAverageResponse(ctx, params.MonitorId.String(), 24)
	if err != nil {
		return oas.SVGBadge{Data: strings.NewReader(renderBadge("Response", "N/A", "#999"))}, nil //nolint:nilerr
	}

	value := fmt.Sprintf("%.0fms", resp)
	return oas.SVGBadge{Data: strings.NewReader(renderBadge("Response", value, "#4c1"))}, nil
}

func (h *Handler) GetCertExpiryBadge(_ context.Context, _ oas.GetCertExpiryBadgeParams) (oas.SVGBadge, error) {
	return oas.SVGBadge{Data: strings.NewReader(renderBadge("Cert Exp.", "N/A", "#999"))}, nil
}

func renderBadge(label, value, color string) string {
	labelWidth := len(label)*7 + 10
	valueWidth := len(value)*7 + 10
	totalWidth := labelWidth + valueWidth

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="%d" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#b)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`,
		totalWidth, totalWidth,
		labelWidth, labelWidth, valueWidth, color, totalWidth,
		labelWidth/2, label, labelWidth/2, label,
		labelWidth+valueWidth/2, value, labelWidth+valueWidth/2, value,
	)
}
