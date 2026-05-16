package tag

type Tag struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Color string `db:"color"`
}

type MonitorTag struct {
	ID        string `db:"id"`
	MonitorID string `db:"monitor_id"`
	TagID     string `db:"tag_id"`
	Value     string `db:"value"`
}
