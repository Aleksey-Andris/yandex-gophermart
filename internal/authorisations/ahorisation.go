package authorisations

type Auth struct {
	ID       int64  `db:"id"`
	Login    string `db:"login" json:"login"`
	Password string `db:"password_hash" json:"password"`
}
