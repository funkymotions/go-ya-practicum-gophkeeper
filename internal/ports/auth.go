package ports

type AuthService interface {
	Authenticate(username, password string) (string, error)
	Register(username, password string) (string, error)
}
