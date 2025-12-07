package session

import (
	"sync"

	"github.com/shaba5h/tgb"
)

type Session struct {
	mu       sync.Mutex
	id       string
	data     map[string]any
	modified bool
}

func (s *Session) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	s.modified = true
}

func (s *Session) Get(key string) any {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.data[key]
	if !ok {
		return nil
	}

	return val
}

func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	s.modified = true
}

func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]any)
	s.modified = true
}

func (s *Session) Empty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.data) == 0
}

func FromContext(ctx *tgb.Context) *Session {
	if val := ctx.Get(contextKey{}); val != nil {
		return val.(*Session)
	}
	return nil
}
