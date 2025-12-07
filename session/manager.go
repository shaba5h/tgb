package session

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shaba5h/tgb"
	"github.com/shaba5h/tgb/logger"
)

type contextKey struct{}

type KeyFunc func(ctx *tgb.Context) (string, error)

type Manager struct {
	store   Store
	ttl     time.Duration
	keyFunc KeyFunc
}

type ManagerOption func(*Manager)

func WithTTL(ttl time.Duration) ManagerOption {
	return func(m *Manager) {
		m.ttl = ttl
	}
}

func WithKeyFunc(keyFunc KeyFunc) ManagerOption {
	return func(m *Manager) {
		m.keyFunc = keyFunc
	}
}

func defaultKeyFunc(ctx *tgb.Context) (string, error) {
	if ctx.User() == nil {
		return "", fmt.Errorf("failed to build session key: user not found")
	}
	return fmt.Sprintf("tg:session:%d", ctx.User().ID), nil
}

func NewManager(store Store, options ...ManagerOption) *Manager {
	m := &Manager{
		store:   store,
		ttl:     24 * time.Hour,
		keyFunc: defaultKeyFunc,
	}
	for _, opt := range options {
		opt(m)
	}
	return m
}

func (m *Manager) Middleware() tgb.Middleware {
	return func(next tgb.Handler) tgb.Handler {
		return func(ctx *tgb.Context) error {
			logger := logger.FromContext(ctx)

			data := make(map[string]any)

			key, keyErr := m.keyFunc(ctx)
			if keyErr != nil {
				if logger != nil {
					logger.Warn("failed to build session key", "error", keyErr)
				}
			} else {
				raw, err := m.store.Get(ctx.Ctx(), key)
				if err != nil && err != ErrNotFound {
					return fmt.Errorf("failed to get session with key %s: %w", key, err)
				}

				if len(raw) > 0 {
					if err := json.Unmarshal(raw, &data); err != nil {
						if logger != nil {
							logger.Warn("failed to unmarshal session", "error", err, "key", key)
						}
					}
				}
			}

			session := &Session{
				id:   key,
				data: data,
			}

			ctx.Set(contextKey{}, session)

			if err := next(ctx); err != nil {
				return err
			}

			if keyErr == nil {
				session.mu.Lock()
				defer session.mu.Unlock()

				if session.modified {
					if session.Empty() {
						if err := m.store.Del(ctx.Ctx(), session.id); err != nil {
							if logger != nil {
								logger.Warn("failed to delete session", "error", err, "key", session.id)
							}
						}
					} else {
						raw, err := json.Marshal(session.data)
						if err != nil {
							if logger != nil {
								logger.Warn("failed to marshal session", "error", err, "key", session.id)
							}
						}
						if err := m.store.Set(ctx.Ctx(), session.id, raw, m.ttl); err != nil {
							return fmt.Errorf("failed to set session: %w", err)
						}
					}
				}
			}

			return nil
		}
	}
}
