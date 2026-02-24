package ports

import (
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/repository"
)

type TypesService interface {
	ReadAllTypes() ([]*model.Type, error)
}

type TypesRepository interface {
	repository.Repository[model.Type]
}
