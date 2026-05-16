package notification

type Notification struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	UserID string `db:"user_id"`
	Config string `db:"config"`
	Active bool   `db:"active"`
}

type MonitorNotification struct {
	MonitorID      string `db:"monitor_id"`
	NotificationID string `db:"notification_id"`
}
