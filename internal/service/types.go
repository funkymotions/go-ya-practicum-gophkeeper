package service

import (
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"go.uber.org/zap"
)

var _ ports.TypesService = (*typesService)(nil)

type typesService struct {
	repository ports.TypesRepository
	logger     *zap.SugaredLogger
}

func NewTypesService(
	r ports.TypesRepository,
	logger *zap.SugaredLogger,
) *typesService {
	return &typesService{
		repository: r,
		logger:     logger,
	}
}

func (s *typesService) ReadAllTypes() ([]*model.Type, error) {
	types, err := s.repository.ReadAll()
	if err != nil {
		s.logger.Error(err)
		return nil, apperror.TypesReadError
	}

	return types, nil
}
