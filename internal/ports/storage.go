package ports

import (
	"context"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/repository"
)

type ClientStorageService interface {
	SaveBlock(ctx context.Context, b model.Block, errChan chan error)
	ListBlockTypes(ctx context.Context, typesChan chan []*model.Type, errChan chan error)
	StartBlockStream(ctx context.Context, blocksChan chan []*model.Block, errChan chan error)
	Ping(ctx context.Context) error
	FileHandler
}

type ClientSubscriber interface {
	Subscribe() error
	Unsubscribe() error
}

type FileHandler interface {
	UnloadToFile() error
	StartupFromFile() error
	SyncOfflineBlocks(context.Context) error
}

type StorageService interface {
	SaveDataBlock(userID int, block *model.Block) (*model.Block, error)
	ListDataBlocks(userID int) ([]*model.Block, error)
}

type StorageRepository interface {
	repository.Repository[model.Block]
	ReadByUserID(userID int) ([]*model.Block, error)
}
