package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shaba5h/tgb"
)

type KeyFunc[C SessionContextProvider[T], T Session] func(ctx C) (string, error)

type Manager[C SessionContextProvider[T], T Session] struct {
	store   Store
	ttl     time.Duration
	keyFunc KeyFunc[C, T]
	factory func() T
}

type ManagerOption[C SessionContextProvider[T], T Session] func(*Manager[C, T])

func WithTTL[C SessionContextProvider[T], T Session](ttl time.Duration) ManagerOption[C, T] {
	return func(m *Manager[C, T]) {
		m.ttl = ttl
	}
}

func WithKeyFunc[C SessionContextProvider[T], T Session](keyFunc KeyFunc[C, T]) ManagerOption[C, T] {
	return func(m *Manager[C, T]) {
		m.keyFunc = keyFunc
	}
}

func defaultKeyFunc[C SessionContextProvider[T], T Session](ctx C) (string, error) {
	if ctx.User() == nil {
		return "", fmt.Errorf("failed to build session key: user not found")
	}
	return fmt.Sprintf("tg:session:%d", ctx.User().ID), nil
}

func NewManager[C SessionContextProvider[T], T Session](store Store, factory func() T, options ...ManagerOption[C, T]) *Manager[C, T] {
	m := &Manager[C, T]{
		store:   store,
		ttl:     24 * time.Hour,
		keyFunc: defaultKeyFunc[C, T],
		factory: factory,
	}
	for _, opt := range options {
		opt(m)
	}
	return m
}

func (m *Manager[C, T]) Middleware() tgb.Middleware[C] {
	return func(next tgb.Handler[C]) tgb.Handler[C] {
		return func(ctx C) error {
			key, err := m.keyFunc(ctx)
			if err != nil {
				return fmt.Errorf("failed to build session key: %w", err)
			}

			raw, err := m.store.Get(ctx.Context(), key)
			if err != nil && !errors.Is(err, ErrNotFound) {
				return fmt.Errorf("failed to get session with key %s: %w", key, err)
			}

			var session T
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &session); err != nil {
					return fmt.Errorf("failed to unmarshal session with key %s: %w", key, err)
				}
			} else {
				session = m.factory()
			}

			ctx.SetSession(session)

			if err := next(ctx); err != nil {
				return err
			}

			session = ctx.GetSession()

			if session.Empty() {
				if err := m.store.Del(ctx.Context(), key); err != nil {
					return fmt.Errorf("failed to delete session: %w", err)
				}
			} else if session.Modified() {
				raw, err := json.Marshal(session)
				if err != nil {
					return fmt.Errorf("failed to marshal session: %w", err)
				}

				if err := m.store.Set(ctx.Context(), key, raw, m.ttl); err != nil {
					return fmt.Errorf("failed to set session: %w", err)
				}
			}

			return nil
		}
	}
}
