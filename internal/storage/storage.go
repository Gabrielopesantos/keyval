package storage

import (
	"errors"
	"fmt"
	"github.com/gabrielopesantos/keyval/internal/item"
	"log/slog"
	"sync"
)

type Manager interface {
	Get(key string) (*item.Item, error)
	Add(item *item.Item) error
	Delete(key string) error
}

type SyncMapStorageManager struct {
	storageMap        *sync.Map
	TTLCleanupEnabled bool
	// Stats struct
	// Dict with TTLs?
	logger *slog.Logger
}

func NewSyncMapStorage(TTLCleanupEnabled bool, logger *slog.Logger) *SyncMapStorageManager {
	return &SyncMapStorageManager{
		storageMap:        &sync.Map{},
		TTLCleanupEnabled: TTLCleanupEnabled,
		logger:            logger,
	}
}

func (s *SyncMapStorageManager) Get(key string) (*item.Item, error) {
	value, ok := s.storageMap.Load(key)
	if !ok {
		return nil, fmt.Errorf("provided key not found: '%s'", key)
	}

	itemValue, ok := value.(*item.Item)
	if !ok {
		s.logger.Error("unexpected condition evaluated: item should be castable to *item.Item") // TODO: Complete description
		// Should be a different Error
		return nil, errors.New("could not parse stored key")
	}

	return itemValue, nil
}

func (s *SyncMapStorageManager) Add(item *item.Item) error {
	if _, ok := s.storageMap.Load(item.Key); ok {
		// NOTE: Create an error for this
		return fmt.Errorf("provided key has already been stored: '%s'", item.Key)
	}
	s.storageMap.Store(item.Key, item)
	return nil
}

func (s *SyncMapStorageManager) Delete(key string) error {
	s.storageMap.Delete(key) // No confirmation?
	return nil
}

//func (s *SyncMapStorageManager) FlushAll() {
//	s.storageMap = &sync.Map{}
//}
