package stats

type StatMinutely struct {
	ID         string  `db:"id"`
	MonitorID  string  `db:"monitor_id"`
	Timestamp  int64   `db:"timestamp"`
	Latency    float64 `db:"latency"`
	LatencyMin int64   `db:"latency_min"`
	LatencyMax int64   `db:"latency_max"`
	Up         int     `db:"up"`
	Down       int     `db:"down"`
	Status     int     `db:"status"`
}

type StatHourly struct {
	ID         string  `db:"id"`
	MonitorID  string  `db:"monitor_id"`
	Timestamp  int64   `db:"timestamp"`
	Latency    float64 `db:"latency"`
	LatencyMin int64   `db:"latency_min"`
	LatencyMax int64   `db:"latency_max"`
	Up         int     `db:"up"`
	Down       int     `db:"down"`
	Status     int     `db:"status"`
}

type StatDaily struct {
	ID         string  `db:"id"`
	MonitorID  string  `db:"monitor_id"`
	Timestamp  int64   `db:"timestamp"`
	Latency    float64 `db:"latency"`
	LatencyMin int64   `db:"latency_min"`
	LatencyMax int64   `db:"latency_max"`
	Up         int     `db:"up"`
	Down       int     `db:"down"`
	Status     int     `db:"status"`
}
