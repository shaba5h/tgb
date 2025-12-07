package store

import (
	"context"
	"sync"
	"time"

	"github.com/shaba5h/tgb/session"
)

type memorySessionStoreItem struct {
	value     []byte
	expiresAt time.Time
}

type MemorySessionStore struct {
	store    sync.Map
	stopChan chan struct{}
}

var _ session.Store = (*MemorySessionStore)(nil)

func NewMemorySessionStore(cleanupInterval time.Duration) *MemorySessionStore {
	m := &MemorySessionStore{
		stopChan: make(chan struct{}),
	}
	if cleanupInterval > 0 {
		go m.gcLoop(cleanupInterval)
	}
	return m
}

func (m *MemorySessionStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.store.Load(key)
	if !ok {
		return nil, session.ErrNotFound
	}

	it := val.(memorySessionStoreItem)
	if time.Now().After(it.expiresAt) {
		m.store.Delete(key)
		return nil, session.ErrNotFound
	}

	return it.value, nil
}

func (m *MemorySessionStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.store.Store(key, memorySessionStoreItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
	return nil
}

func (m *MemorySessionStore) Del(ctx context.Context, key string) error {
	m.store.Delete(key)
	return nil
}

func (m *MemorySessionStore) Stop() {
	close(m.stopChan)
}

func (m *MemorySessionStore) gcLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

func (m *MemorySessionStore) cleanup() {
	now := time.Now()
	m.store.Range(func(key, value any) bool {
		item := value.(memorySessionStoreItem)
		if now.After(item.expiresAt) {
			m.store.Delete(key)
		}
		return true
	})
}
