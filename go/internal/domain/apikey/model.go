package apikey

import "time"

type APIKey struct {
	ID        string     `db:"id"`
	Key       string     `db:"key"`
	Name      string     `db:"name"`
	UserID    string     `db:"user_id"`
	CreatedAt time.Time  `db:"created_date"`
	Active    bool       `db:"active"`
	Expires   *time.Time `db:"expires"`
}
