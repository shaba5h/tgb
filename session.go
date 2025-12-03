package tgb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type SessionStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
}

var ErrSessionNotFound = errors.New("session not found")

type sessionKey struct{}

type MemorySessionStore struct {
	store sync.Map
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{}
}

func (m *MemorySessionStore) Get(ctx context.Context, key string) ([]byte, error) {
	if value, ok := m.store.Load(key); ok {
		return value.([]byte), nil
	}
	return nil, ErrSessionNotFound
}

func (m *MemorySessionStore) Set(ctx context.Context, key string, value []byte) error {
	m.store.Store(key, value)
	return nil
}

func (m *MemorySessionStore) Delete(ctx context.Context, key string) error {
	m.store.Delete(key)
	return nil
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

type SessionManager[T any] struct {
	store      SessionStore
	keyBuilder KeyBuilder
}

type SessionManagerOptions struct {
	store      SessionStore
	keyBuilder KeyBuilder
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

func NewSessionManager[T any](options ...SessionManagerOption) *SessionManager[T] {
	opts := SessionManagerOptions{
		store:      NewMemorySessionStore(),
		keyBuilder: NewUserKeyBuilder(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return &SessionManager[T]{
		store:      opts.store,
		keyBuilder: opts.keyBuilder,
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
			if val != nil {
				data, err := json.Marshal(val)
				if err != nil {
					return fmt.Errorf("failed to marshal session: %w", err)
				}
				return s.store.Set(ctx.Ctx(), key, data)
			}

			return nil
		}
	}
}

func Session[T any](ctx *Context) *T {
	val := ctx.Get(sessionKey{})
	if val == nil {
		return new(T)
	}
	return val.(*T)
}
