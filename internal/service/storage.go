package service

import (
	"errors"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"go.uber.org/zap"
)

type storageService struct {
	storageRepository   ports.StorageRepository
	typesRepository     ports.TypesRepository
	subscriptionService ports.SubscriptionService
	logger              *zap.SugaredLogger
}

type StorageServiceArgs struct {
	StorageRepository   ports.StorageRepository
	TypesRepository     ports.TypesRepository
	SubscriptionService ports.SubscriptionService
	Logger              *zap.SugaredLogger
}

var _ ports.StorageService = (*storageService)(nil)

func NewStorageService(args StorageServiceArgs) *storageService {
	return &storageService{
		storageRepository:   args.StorageRepository,
		subscriptionService: args.SubscriptionService,
		typesRepository:     args.TypesRepository,
		logger:              args.Logger,
	}
}

func (s *storageService) SaveDataBlock(userID int, in *model.Block) (*model.Block, error) {
	block, err := s.storageRepository.Create(in)
	if err != nil {
		s.logger.Error(err)
		return nil, apperror.StorageCreateBlockError
	}

	blocks, err := s.ListDataBlocks(userID)
	if err != nil {
		s.logger.Error(err)
		return nil, apperror.StorageListDataBlockError
	}

	s.subscriptionService.NotifySubscribers(userID, blocks)

	s.logger.Infow("Data block saved successfully", "blockID", block.ID, "userID", userID)

	return block, nil
}

func (s *storageService) ListDataBlocks(userID int) ([]*model.Block, error) {
	blocks, err := s.storageRepository.ReadByUserID(userID)
	if err != nil && errors.Is(err, apperror.DBErrorNoRows) {
		return []*model.Block{}, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, apperror.StorageListDataBlockError
	}

	s.logger.Infow("Data blocks fetched successfully", "userID", userID)

	return blocks, nil
}
