package user

type User struct {
	ID             string `db:"id"`
	Username       string `db:"username"`
	Password       string `db:"password"`
	Active         bool   `db:"active"`
	TwoFAStatus    bool   `db:"twofa_status"`
	TwoFASecret    string `db:"twofa_secret"`
	TwoFALastToken string `db:"twofa_last_token"`
}
