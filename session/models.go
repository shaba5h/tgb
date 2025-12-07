package session

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("session not found")

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}
