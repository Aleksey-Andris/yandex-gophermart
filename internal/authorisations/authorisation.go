package authorisations

type Auth struct {
	ID       int64
	Login    string `json:"login"`
	Password string `json:"password"`
}
