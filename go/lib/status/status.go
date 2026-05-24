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
		return "down"
	case Up:
		return "up"
	case Pending:
		return "pending"
	case Maintenance:
		return "maintenance"
	case Degraded:
		return "degraded"
	default:
		return "down"
	}
}

func FromString(s string) Status {
	switch s {
	case "down":
		return Down
	case "up":
		return Up
	case "pending":
		return Pending
	case "maintenance":
		return Maintenance
	case "degraded":
		return Degraded
	default:
		return Down
	}
}
