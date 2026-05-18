package monitor

import (
	"context"
	"time"

	"github.com/koblas/besops/lib/status"
)

type CheckResult struct {
	Status       status.Status
	Latency      int64
	Message      string
	ResponseBody []byte
	CertInfo     *CertInfo
}

type CertInfo struct {
	Valid    bool
	Issuer   string
	Subject  string
	DaysLeft int
	ExpiryAt string
}

// Config holds the common check configuration shared by all monitor types,
// plus type-specific sub-configs. Each checker reads only its own sub-config.
type Config struct {
	ID            string
	Type          string
	Name          string
	URL           string
	Hostname      string
	Port          int
	Interval      time.Duration
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	IgnoreTLS     bool
	Keyword       string
	KeywordType   string
	JsonPath      string
	ExpectedValue string
	ProxyID     *string
	PushToken   string
	Tags        []string
	GroupTagIDs []string

	HTTP  HTTPConfig
	DNS   DNSConfig
	MQTT  MQTTConfig
	GRPC  GRPCConfig
	SMTP  SMTPConfig
	Redis RedisConfig
}

type HTTPConfig struct {
	Method              string
	Headers             map[string]string
	Body                string
	AcceptedStatusCodes []string
	BasicAuthUser       string
	BasicAuthPass       string
}

type DNSConfig struct {
	ResolveType string
}

type MQTTConfig struct {
	Topic          string
	Username       string
	Password       string
	SuccessMessage string
	CheckType      string
	WebsocketPath  string
}

type GRPCConfig struct {
	URL         string
	ServiceName string
	Method      string
	Body        string
	Protobuf    string
	EnableTLS   bool
}

type SMTPConfig struct {
	Security string
}

type RedisConfig struct {
	ConnectionString string
}

type ConditionGroup struct {
	Operator   string
	Conditions []Condition
}

type Condition struct {
	Variable string
	Operator string
	Value    string
}

// Checker is the interface that all monitor types must implement.
type Checker interface {
	Type() string
	Check(ctx context.Context, cfg *Config) (CheckResult, error)
}

// ConditionSupporter is optionally implemented by Checkers that support condition-based evaluation.
type ConditionSupporter interface {
	SupportsConditions() bool
	ConditionVariables() []ConditionVariable
}

// CustomStatusChecker is optionally implemented by Checkers that allow non-UP success statuses.
type CustomStatusChecker interface {
	AllowCustomStatus() bool
}

// AggregateChecker is implemented by checkers that derive status from other monitors
// (e.g., groups). These checkers skip retry logic and timeout configuration since
// their result is deterministic at any point in time.
type AggregateChecker interface {
	IsAggregate()
}

type ConditionVariable struct {
	Name      string
	Operators []string
}
