package telemetry

import "context"

// MonitorInfo provides identity attributes for a monitor, used to label telemetry data.
type MonitorInfo interface {
	MonitorID() string
	MonitorName() string
	MonitorType() string
	GroupName() string
	Tags() []string
}

// Observer receives every check result for metrics/telemetry export.
type Observer interface {
	Record(ctx context.Context, monitor MonitorInfo, up bool, latencyMs int64)
}
