package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
)

// Session сессия авторизации пользователя. У пользователя может быть больше одной сессии
type Session struct {
	UID       string
	SessionID string
	CreatedAt time.Time
}

type SessionOption func(session *Session)

func New(uid string, opts ...SessionOption) *Session {
	session := &Session{
		UID:       uid,
		SessionID: random.SessionID(),
		CreatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(session)
	}
	return session
}

func NewRandomSession(uid string) *Session {
	return New(uid, WithSessionID(random.SessionID()), WithCreatedAt(time.Now()))
}

func WithSessionID(sessionID string) SessionOption {
	return func(o *Session) {
		o.SessionID = sessionID
	}
}

func WithCreatedAt(createdAt time.Time) SessionOption {
	return func(o *Session) {
		o.CreatedAt = createdAt
	}
}
