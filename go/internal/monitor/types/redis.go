package types

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type RedisChecker struct{}

func (c *RedisChecker) Type() string { return "redis" }

func (c *RedisChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	dsn := cfg.Redis.ConnectionString
	if dsn == "" {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "database connection string is required",
		}, nil
	}

	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("invalid connection string: %v", err),
		}, nil
	}

	if cfg.IgnoreTLS && opts.TLSConfig != nil {
		opts.TLSConfig.InsecureSkipVerify = true //nolint:gosec // user-configurable
	} else if cfg.IgnoreTLS {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-configurable
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	opts.DialTimeout = timeout
	opts.ReadTimeout = timeout

	client := redis.NewClient(opts)
	defer client.Close()

	start := time.Now()
	result, err := client.Ping(ctx).Result()
	ping := time.Since(start).Milliseconds()

	if err != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("redis ping failed: %v", err),
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    ping,
		Message: result,
	}, nil
}
