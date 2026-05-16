package stats

type StatMinutely struct {
	ID        string  `db:"id"`
	MonitorID string  `db:"monitor_id"`
	Timestamp int64   `db:"timestamp"`
	Ping      float64 `db:"ping"`
	PingMin   int64   `db:"ping_min"`
	PingMax   int64   `db:"ping_max"`
	Up        int     `db:"up"`
	Down      int     `db:"down"`
	Status    int     `db:"status"`
}

type StatHourly struct {
	ID        string  `db:"id"`
	MonitorID string  `db:"monitor_id"`
	Timestamp int64   `db:"timestamp"`
	Ping      float64 `db:"ping"`
	PingMin   int64   `db:"ping_min"`
	PingMax   int64   `db:"ping_max"`
	Up        int     `db:"up"`
	Down      int     `db:"down"`
	Status    int     `db:"status"`
}

type StatDaily struct {
	ID        string  `db:"id"`
	MonitorID string  `db:"monitor_id"`
	Timestamp int64   `db:"timestamp"`
	Ping      float64 `db:"ping"`
	PingMin   int64   `db:"ping_min"`
	PingMax   int64   `db:"ping_max"`
	Up        int     `db:"up"`
	Down      int     `db:"down"`
	Status    int     `db:"status"`
}
