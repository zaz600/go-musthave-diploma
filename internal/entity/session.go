package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type Session struct {
	UID       string
	SessionID string
	CreatedAt int64
}

func NewSession(uid string) *Session {
	return &Session{
		UID:       uid,
		SessionID: random.String(32),
		CreatedAt: time.Now().UnixMilli(),
	}
}
