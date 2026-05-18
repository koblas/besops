package status

type Status int

const (
	Down        Status = 0
	Up          Status = 1
	Pending     Status = 2
	Maintenance Status = 3
	Degraded    Status = 4
)

func (s Status) String() string {
	switch s {
	case Down:
		return "DOWN"
	case Up:
		return "UP"
	case Pending:
		return "PENDING"
	case Maintenance:
		return "MAINTENANCE"
	case Degraded:
		return "DEGRADED"
	default:
		return "UNKNOWN"
	}
}
