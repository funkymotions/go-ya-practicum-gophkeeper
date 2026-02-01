package ports

import "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"

type StorageService interface {
	SaveDataBlock(userID int, block *model.Block) (*model.Block, error)
	ListDataBlocks(userID int) ([]*model.Block, error)
	GetBlockTypes() ([]*model.Type, error)
}

type StorageRepository interface {
	CreateBlock(block *model.Block) (*model.Block, error)
	ReadUserBlocks(userID int) ([]*model.Block, error)
	ReadBlockTypes() ([]*model.Type, error)
}
