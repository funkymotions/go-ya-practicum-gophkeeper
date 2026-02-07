package ports

import (
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

type OfflineBlockReader interface {
	ReadBlocks() []*model.Block
	CountUnsyncedBlocks() int
	ReadUnsyncedBlocks() []*model.Block
}

type OfflineBlockWriter interface {
	SaveUnsyncedBlock(b model.Block)
	SaveSyncedBlock(b model.Block)
	DeleteBlockByData(data []byte)
	ReplaceSyncedBlocks(blocks []*model.Block)
}

type OfflineTypesReader interface {
	ReadTypes() []*model.Type
}

type OfflineTypesWriter interface {
	SaveTypes(types []*model.Type)
}

type OfflineFileHandler interface {
	SaveToFile() error
	RestoreFromFile() error
}

type OfflineAppStateReader interface {
	GetAppState() *types.State
}
