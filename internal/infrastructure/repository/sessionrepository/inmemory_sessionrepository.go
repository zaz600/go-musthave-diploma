package sessionrepository

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemorySessionRepository struct {
	mu sync.RWMutex
	db map[string]*entity.Session
}

func (r *InmemorySessionRepository) AddSession(ctx context.Context, session *entity.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[session.SessionID]; ok {
		return ErrSessionExists
	}
	r.db[session.SessionID] = session
	return nil
}

func (r *InmemorySessionRepository) DelSession(ctx context.Context, sessionID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *InmemorySessionRepository) GetSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if sessionEntity, ok := r.db[sessionID]; ok {
		return sessionEntity, nil
	}
	return nil, ErrSessionNotFound
}

func (r *InmemorySessionRepository) Close() error {
	return nil
}

func NewInmemorySessionRepository() *InmemorySessionRepository {
	return &InmemorySessionRepository{
		mu: sync.RWMutex{},
		db: make(map[string]*entity.Session, 100),
	}
}
