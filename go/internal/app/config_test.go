package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	testArgs = []string{}
	defer func() { testArgs = nil }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "", cfg.Host)
	assert.Equal(t, 3001, cfg.Port)
	assert.Equal(t, "sqlite://./data/besops.db", cfg.DatabaseURL)
	assert.False(t, cfg.SSLEnabled)
	assert.Equal(t, 180, cfg.KeepDataPeriodDays)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "", cfg.JWTSecret)
	assert.False(t, cfg.Bootstrap)
}

func TestConfigEnvOverridesDefaults(t *testing.T) {
	testArgs = []string{}
	defer func() { testArgs = nil }()

	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "sqlite:///tmp/test.db")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("KEEP_DATA_PERIOD_DAYS", "30")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "sqlite:///tmp/test.db", cfg.DatabaseURL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "env-secret", cfg.JWTSecret)
	assert.Equal(t, 30, cfg.KeepDataPeriodDays)
}

func TestConfigFlagsOverrideEnv(t *testing.T) {
	testArgs = []string{"--port", "4000", "--log-level", "warn"}
	defer func() { testArgs = nil }()

	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("JWT_SECRET", "env-secret")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Flags should win over env
	assert.Equal(t, 4000, cfg.Port)
	assert.Equal(t, "warn", cfg.LogLevel)

	// Env should still apply for non-flagged values
	assert.Equal(t, "env-secret", cfg.JWTSecret)
}

func TestConfigFlagsOverrideDefaults(t *testing.T) {
	testArgs = []string{"--port", "5555", "--database-url", "sqlite:///tmp/flag.db", "--bootstrap"}
	defer func() { testArgs = nil }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 5555, cfg.Port)
	assert.Equal(t, "sqlite:///tmp/flag.db", cfg.DatabaseURL)
	assert.True(t, cfg.Bootstrap)

	// Other fields remain at defaults
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 180, cfg.KeepDataPeriodDays)
}

func TestConfigEnvDoesNotOverrideFlaggedValues(t *testing.T) {
	testArgs = []string{"--port", "7777"}
	defer func() { testArgs = nil }()

	t.Setenv("PORT", "8888")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 7777, cfg.Port)
}

func TestConfigUnsetEnvKeepsDefault(t *testing.T) {
	testArgs = []string{}
	defer func() { testArgs = nil }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 3001, cfg.Port)
	assert.Equal(t, "sqlite://./data/besops.db", cfg.DatabaseURL)
}
