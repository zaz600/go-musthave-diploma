package sessionservice

import (
	"context"
	"fmt"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
)

type SessionService interface {
	NewSession(ctx context.Context, userID string) (*entity.Session, error)
	Get(ctx context.Context, sessionID string) (*entity.Session, error)
}

type Service struct {
	sessionRepository sessionrepository.SessionRepository
}

func (s Service) NewSession(ctx context.Context, userID string) (*entity.Session, error) {
	// TODO проверить на пустоту userID или сделать констрейнт
	session := entity.NewSession(userID)
	if err := s.sessionRepository.AddSession(ctx, session); err != nil {
		return nil, fmt.Errorf("error creating user session: %w", err)
	}
	return session, nil
}

func (s Service) Get(ctx context.Context, sessionID string) (*entity.Session, error) {
	session, err := s.sessionRepository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("error get session: %w", err)
	}
	return session, nil
}

func NewService(sessionRepository sessionrepository.SessionRepository) *Service {
	return &Service{
		sessionRepository: sessionRepository,
	}
}
