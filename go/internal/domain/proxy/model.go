package proxy

type Proxy struct {
	ID       string
	UserID   string
	Protocol string
	Host     string
	Port     int64
	Auth     bool
	Username string
	Password string
	Active   bool
	Default  bool
}
