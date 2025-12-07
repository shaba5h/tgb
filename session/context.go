package session

import "github.com/shaba5h/tgb"

type Session interface {
	Empty() bool
	Modified() bool
}

type SessionProvider[T Session] interface {
	GetSession() T
	SetSession(T)
}

type SessionContextProvider[T Session] interface {
	tgb.ContextProvider
	SessionProvider[T]
}

type SessionContext[T Session] struct {
	Session T
}

func NewSessionContext[T Session]() *SessionContext[T] {
	return &SessionContext[T]{}
}

func (s *SessionContext[T]) GetSession() T {
	return s.Session
}

func (s *SessionContext[T]) SetSession(session T) {
	s.Session = session
}
