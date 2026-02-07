package ports

type ClientAuthService interface {
	Authenticate(username, password string) (string, error)
	Register(username, password string) (string, error)
}

type AuthService interface {
	Authenticate(username, password string) (string, error)
	Register(username, password string) (string, error)
}
