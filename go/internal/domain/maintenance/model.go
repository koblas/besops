package maintenance

import "time"

type Maintenance struct {
	ID              string     `db:"id"`
	Title           string     `db:"title"`
	Description     string     `db:"description"`
	UserID          string     `db:"user_id"`
	Active          bool       `db:"active"`
	Strategy        string     `db:"strategy"`
	StartDate       *time.Time `db:"start_date"`
	EndDate         *time.Time `db:"end_date"`
	StartTime       string     `db:"start_time"`
	EndTime         string     `db:"end_time"`
	Weekdays        string     `db:"weekdays"`
	DaysOfMonth     string     `db:"days_of_month"`
	IntervalDay     int        `db:"interval_day"`
	CronExpression  string     `db:"cron"`
	DurationMinutes int        `db:"duration_minutes"`
	TimezoneOption  string     `db:"timezone_option"`
}

type MonitorLink struct {
	MaintenanceID string `db:"maintenance_id"`
	MonitorID     string `db:"monitor_id"`
}
