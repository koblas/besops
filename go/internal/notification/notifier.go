package notification

import "context"

type MonitorInfo struct {
	ID       int64
	Name     string
	URL      string
	Type     string
	Hostname string
	Port     int
	Tags     []Tag
}

type Tag struct {
	ID    int64
	Name  string
	Color string
	Value string
}

type HeartbeatInfo struct {
	Status   int
	Time     string
	Message  string
	Latency  int64
	Duration int64
	Timezone string
}

// Notifier is the interface that all notification providers must implement.
type Notifier interface {
	Name() string
	Send(ctx context.Context, config map[string]any, msg string, monitor *MonitorInfo, heartbeat *HeartbeatInfo) error
}

// Validator is optionally implemented by Notifiers that can validate their config at creation time.
type Validator interface {
	ValidateConfig(config map[string]any) error
}
