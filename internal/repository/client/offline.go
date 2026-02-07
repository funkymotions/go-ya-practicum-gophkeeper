package client

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

const storageFileName = "storage.json"

type storage struct {
	AppState *types.State   `json:"state"`
	Synced   []*model.Block `json:"synced"`
	Unsynced []*model.Block `json:"unsynced"`
	Types    []*model.Type  `json:"types"`
}

type offlineRepository struct {
	storage storage
}

type offlineRepoInterface interface {
	ports.OfflineBlockReader
	ports.OfflineBlockWriter
	ports.OfflineTypesReader
	ports.OfflineTypesWriter
	ports.OfflineFileHandler
	ports.OfflineAppStateReader
}

var _ offlineRepoInterface = (*offlineRepository)(nil)

func NewOfflineRepository(clientState *types.State) *offlineRepository {
	return &offlineRepository{
		storage: storage{
			AppState: clientState,
			Synced:   make([]*model.Block, 0),
			Unsynced: make([]*model.Block, 0),
			Types:    make([]*model.Type, 0),
		},
	}
}

func (s *offlineRepository) SaveUnsyncedBlock(b model.Block) {
	s.storage.Unsynced = append(s.storage.Unsynced, &b)
}

func (s *offlineRepository) SaveSyncedBlock(b model.Block) {
	s.storage.Synced = append(s.storage.Synced, &b)
}

func (s *offlineRepository) ReplaceSyncedBlocks(blocks []*model.Block) {
	s.storage.Synced = blocks
}

func (s *offlineRepository) DeleteBlockByData(data []byte) {
	for i, b := range s.storage.Unsynced {
		if bytes.Equal(b.Data, data) {
			s.storage.Unsynced = append(s.storage.Unsynced[:i], s.storage.Unsynced[i+1:]...)
			return
		}
	}
}

func (s *offlineRepository) ReadBlocks() []*model.Block {
	blocks := make([]*model.Block, 0, len(s.storage.Synced)+len(s.storage.Unsynced))
	blocks = append(blocks, s.storage.Synced...)
	blocks = append(blocks, s.storage.Unsynced...)

	return blocks
}

func (s *offlineRepository) ReadUnsyncedBlocks() []*model.Block {
	return s.storage.Unsynced
}

func (s *offlineRepository) CountUnsyncedBlocks() int {
	return len(s.storage.Unsynced)
}

func (s *offlineRepository) ReadTypes() []*model.Type {
	return s.storage.Types
}

func (s *offlineRepository) SaveTypes(types []*model.Type) {
	s.storage.Types = types
}

func (s *offlineRepository) GetAppState() *types.State {
	return s.storage.AppState
}

func (s *offlineRepository) SaveToFile() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := filepath.Dir(execPath)
	content, err := json.MarshalIndent(s.storage, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(execDir, storageFileName), content, 0644); err != nil {
		return err
	}

	return nil
}

func (s *offlineRepository) RestoreFromFile() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := filepath.Dir(path)
	filePath := filepath.Join(execDir, storageFileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &s.storage); err != nil {
		return err
	}

	return nil
}
