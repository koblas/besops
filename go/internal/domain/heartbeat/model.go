package heartbeat

import "time"

type RFC3339Time time.Time

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).UTC().Format(time.RFC3339) + `"`), nil
}

type Heartbeat struct {
	ID        string      `db:"id" json:"id"`
	MonitorID string      `db:"monitor_id" json:"monitorId"`
	Status    int         `db:"status" json:"status"`
	Time      RFC3339Time `db:"time" json:"time"`
	Msg       string      `db:"msg" json:"msg,omitempty"`
	Latency   *int64      `db:"latency" json:"latency,omitempty"`
	Important bool        `db:"important" json:"important,omitempty"`
	Duration  int64       `db:"duration" json:"duration,omitempty"`
	Retries   int         `db:"retries" json:"retries,omitempty"`
	Response  []byte      `db:"response" json:"-"`
}
