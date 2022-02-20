package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
)

type Session struct {
	UID       string
	SessionID string
	CreatedAt time.Time
}

func NewSession(uid string) *Session {
	return &Session{
		UID:       uid,
		SessionID: random.String(32),
		CreatedAt: time.Now(),
	}
}
