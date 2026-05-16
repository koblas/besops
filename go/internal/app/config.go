package app

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// testArgs allows tests to override CLI arguments. When nil, os.Args[1:] is used.
var testArgs []string

type Config struct {
	Host               string `koanf:"host"`
	Port               int    `koanf:"port"`
	DatabaseURL        string `koanf:"database_url"`
	SSLEnabled         bool   `koanf:"ssl"`
	SSLCert            string `koanf:"ssl_cert"`
	SSLKey             string `koanf:"ssl_key"`
	DemoMode           bool   `koanf:"demo"`
	TrustProxy         bool   `koanf:"trust_proxy"`
	KeepDataPeriodDays int    `koanf:"keep_data_period_days"`
	Bootstrap          bool   `koanf:"bootstrap"`
	LogLevel           string `koanf:"log_level"`
	JWTSecret          string `koanf:"jwt_secret"`
}

func DefaultConfig() Config {
	return Config{
		Host:               "",
		Port:               3001,
		DatabaseURL:        "sqlite://./data/besops.db",
		SSLEnabled:         false,
		SSLCert:            "",
		SSLKey:             "",
		DemoMode:           false,
		TrustProxy:         false,
		KeepDataPeriodDays: 180,
		Bootstrap:          false,
		LogLevel:           "info",
	}
}

// flagToKoanf maps CLI flag names (kebab-case) to koanf keys (snake_case).
var flagToKoanf = map[string]string{
	"host":                 "host",
	"port":                 "port",
	"database-url":        "database_url",
	"ssl":                  "ssl",
	"ssl-cert":            "ssl_cert",
	"ssl-key":             "ssl_key",
	"demo":                "demo",
	"trust-proxy":         "trust_proxy",
	"keep-data-period-days": "keep_data_period_days",
	"bootstrap":           "bootstrap",
	"log-level":           "log_level",
	"jwt-secret":          "jwt_secret",
}

// registerFlags creates the FlagSet with all recognized flags.
func registerFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("rupert", flag.ContinueOnError)

	fs.String("host", "", "listen host")
	fs.Int("port", 0, "listen port")
	fs.String("database-url", "", "database connection URL")
	fs.Bool("ssl", false, "enable TLS")
	fs.String("ssl-cert", "", "TLS certificate file")
	fs.String("ssl-key", "", "TLS key file")
	fs.Bool("demo", false, "enable demo mode")
	fs.Bool("trust-proxy", false, "trust reverse proxy headers")
	fs.Int("keep-data-period-days", 0, "days to retain heartbeat data")
	fs.Bool("bootstrap", false, "auto-create users on login")
	fs.String("log-level", "", "log level: debug, info, warn, error")
	fs.String("jwt-secret", "", "stable secret for deriving JWT signing key (random per-start if empty)")

	return fs
}

// LoadConfig loads configuration from defaults, environment variables, and CLI flags.
// Priority (highest wins): flags > env > defaults.
func LoadConfig() (Config, error) {
	k := koanf.New(".")

	// 1. Defaults
	if err := k.Load(structs.Provider(DefaultConfig(), "koanf"), nil); err != nil {
		return Config{}, fmt.Errorf("loading defaults: %w", err)
	}

	// 2. Environment variables: no prefix, no nesting — full lowercase name is the key
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(s)
	}), nil); err != nil {
		return Config{}, fmt.Errorf("loading env: %w", err)
	}

	// 3. CLI flags — only those explicitly passed override
	fs := registerFlags()
	args := testArgs
	if args == nil {
		args = os.Args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parsing flags: %w", err)
	}

	fs.Visit(func(f *flag.Flag) {
		koanfKey, ok := flagToKoanf[f.Name]
		if !ok {
			return
		}
		_ = k.Set(koanfKey, f.Value.(flag.Getter).Get())
	})

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshalling config: %w", err)
	}

	return cfg, nil
}

func (c Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
