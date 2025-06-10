package services

import (
	"context"
	"fmt"
	"time"

	"github.com/prompt-ops/pops/internal/domain"
)

type SessionService struct {
	repo domain.SessionRepository
}

func NewSessionService(repo domain.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, connectionID domain.ConnectionID) (*domain.Session, error) {
	session := &domain.Session{
		ID:           domain.SessionID(generateID()),
		ConnectionID: connectionID,
		State:        domain.SessionStateActive,
		History:      []domain.Interaction{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActive:   time.Now(),
	}

	if err := s.repo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error) {
	return s.repo.FindByID(ctx, sessionID)
}

func (s *SessionService) AddInteraction(ctx context.Context, sessionID domain.SessionID, interaction domain.Interaction) error {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.History = append(session.History, interaction)
	session.UpdatedAt = time.Now()
	session.LastActive = time.Now()

	return s.repo.Save(ctx, session)
}

func (s *SessionService) UpdateSessionState(ctx context.Context, sessionID domain.SessionID, state domain.SessionState) error {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.State = state
	session.UpdatedAt = time.Now()

	return s.repo.Save(ctx, session)
}

func (s *SessionService) CloseSession(ctx context.Context, sessionID domain.SessionID) error {
	return s.UpdateSessionState(ctx, sessionID, domain.SessionStateCompleted)
}

func (s *SessionService) GetActiveSessions(ctx context.Context) ([]*domain.Session, error) {
	return s.repo.FindActive(ctx)
}
