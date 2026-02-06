package service

import (
	"errors"
	"fmt"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"go.uber.org/zap"
)

type storageService struct {
	storageRepository   ports.StorageRepository
	subscriptionService ports.SubscriptionService
	logger              *zap.SugaredLogger
}

type StorageServiceArgs struct {
	StorageRepository   ports.StorageRepository
	SubscriptionService ports.SubscriptionService
	Logger              *zap.SugaredLogger
}

var _ ports.StorageService = (*storageService)(nil)

func NewStorageService(args StorageServiceArgs) *storageService {
	return &storageService{
		storageRepository:   args.StorageRepository,
		subscriptionService: args.SubscriptionService,
		logger:              args.Logger,
	}
}

func (s *storageService) SaveDataBlock(userID int, in *model.Block) (*model.Block, error) {
	block, err := s.storageRepository.CreateBlock(in)
	if err != nil {
		s.logger.Error(err)
		return nil, &apperror.StorageCreateBlockError
	}

	blocks, err := s.ListDataBlocks(userID)
	if err != nil {
		return nil, &apperror.StorageListDataBlockError
	}

	s.subscriptionService.NotifySubscribers(userID, blocks)

	return block, nil
}

func (s *storageService) ListDataBlocks(userID int) ([]*model.Block, error) {
	blocks, err := s.storageRepository.ReadUserBlocks(userID)
	fmt.Printf("blocks retrieved from repository: %v\n", err)
	if err != nil && errors.Is(err, apperror.DBErrorNoRows) {
		fmt.Printf("blocks in ListDataBlocks: %v\n", blocks)
		return []*model.Block{}, nil
	}
	if err != nil {
		return nil, &apperror.StorageListDataBlockError
	}

	return blocks, nil
}

func (s *storageService) GetBlockTypes() ([]*model.Type, error) {
	types, err := s.storageRepository.ReadBlockTypes()
	if err != nil {
		return nil, &apperror.StorageReadBlockTypeError
	}

	return types, nil
}
