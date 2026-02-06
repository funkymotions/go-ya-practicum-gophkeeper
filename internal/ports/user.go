package ports

import "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"

type UserRepositoryReader interface {
	ReadUserByID(id int32) (*model.User, error)
	ReadUserByUsername(username string) (*model.User, error)
}

type UserRepositoryWriter interface {
	CreateUser(user *model.User) (*model.User, error)
}
