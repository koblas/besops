package monitor

import "time"

type Monitor struct {
	ID                 string    `db:"id"`
	Name               string    `db:"name"`
	Active             bool      `db:"active"`
	UserID             string    `db:"user_id"`
	Interval           int       `db:"interval"`
	Type               string    `db:"type"`
	Weight             int       `db:"weight"`
	CreatedDate        time.Time `db:"created_date"`
	MaxRetries         int       `db:"maxretries"`
	UpsideDown         bool      `db:"upside_down"`
	RetryInterval      int       `db:"retry_interval"`
	Timeout            int       `db:"timeout"`
	Description        string    `db:"description"`
	ResendInterval     int       `db:"resend_interval"`
	ExpiryNotification bool      `db:"expiry_notification"`
	ConfigJSON         string    `db:"config_json"`
}
