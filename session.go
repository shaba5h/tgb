package tgb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type SessionStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

var ErrSessionNotFound = errors.New("session not found")

type sessionKey struct{}

type memorySessionStoreItem struct {
	value     []byte
	expiresAt time.Time
}

const memorySessionStoreGCInterval = 1 * time.Minute

type MemorySessionStore struct {
	store    sync.Map
	stopChan chan struct{}
}

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
		return nil, ErrSessionNotFound
	}

	it := val.(memorySessionStoreItem)
	if time.Now().After(it.expiresAt) {
		m.store.Delete(key)
		return nil, ErrSessionNotFound
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

func (m *MemorySessionStore) Delete(ctx context.Context, key string) error {
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

type KeyBuilder interface {
	Build(ctx *Context) string
}

type UserKeyBuilder struct{}

func (c UserKeyBuilder) Build(ctx *Context) string {
	if user := ctx.User(); user != nil {
		return fmt.Sprintf("telegram:session:%d", user.ID)
	}
	return ""
}

func NewUserKeyBuilder() KeyBuilder {
	return &UserKeyBuilder{}
}

const defaultSessionManagerTTL = 24 * time.Hour

type SessionManager[T any] struct {
	store      SessionStore
	keyBuilder KeyBuilder
	defaultTTL time.Duration
}

type SessionManagerOptions struct {
	store      SessionStore
	keyBuilder KeyBuilder
	ttl        time.Duration
}

type SessionManagerOption func(*SessionManagerOptions)

func WithStore(store SessionStore) SessionManagerOption {
	return func(options *SessionManagerOptions) {
		options.store = store
	}
}

func WithKeyBuilder(keyBuilder KeyBuilder) SessionManagerOption {
	return func(options *SessionManagerOptions) {
		options.keyBuilder = keyBuilder
	}
}

func WithTTL(ttl time.Duration) SessionManagerOption {
	return func(options *SessionManagerOptions) {
		options.ttl = ttl
	}
}

func NewSessionManager[T any](options ...SessionManagerOption) *SessionManager[T] {
	opts := SessionManagerOptions{
		store:      NewMemorySessionStore(memorySessionStoreGCInterval),
		keyBuilder: NewUserKeyBuilder(),
		ttl:        defaultSessionManagerTTL,
	}
	for _, opt := range options {
		opt(&opts)
	}
	return &SessionManager[T]{
		store:      opts.store,
		keyBuilder: opts.keyBuilder,
		defaultTTL: opts.ttl,
	}
}

func (s *SessionManager[T]) Middleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			key := s.keyBuilder.Build(ctx)
			if key == "" {
				return next(ctx)
			}

			var session T

			data, err := s.store.Get(ctx.Ctx(), key)
			if err != nil && !errors.Is(err, ErrSessionNotFound) {
				// Critical error. Fail fast.
				return fmt.Errorf("session store get error: %w", err)
			}

			if len(data) > 0 {
				if err := json.Unmarshal(data, &session); err != nil {
					return fmt.Errorf("failed to unmarshal session: %w", err)
				}
			}

			ctx.Set(sessionKey{}, &session)

			if err = next(ctx); err != nil {
				return err
			}

			val := ctx.Get(sessionKey{})
			if val == nil {
				// Session was explicitly reset
				return s.store.Delete(ctx.Ctx(), key)
			}

			newData, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal session: %w", err)
			}

			if !bytes.Equal(newData, data) {
				return s.store.Set(ctx.Ctx(), key, newData, s.defaultTTL)
			}

			return nil
		}
	}
}

func GetSession[T any](ctx *Context) *T {
	val := ctx.Get(sessionKey{})
	if val == nil {
		return new(T)
	}
	return val.(*T)
}

func ResetSession(ctx *Context) {
	ctx.Set(sessionKey{}, nil)
}
