package repositories

import (
	"context"
	"fmt"

	"github.com/prompt-ops/pops/internal/domain"
)

type MemorySessionRepository struct {
	sessions map[domain.SessionID]*domain.Session
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[domain.SessionID]*domain.Session),
	}
}

func (r *MemorySessionRepository) Save(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *MemorySessionRepository) FindByID(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error) {
	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

func (r *MemorySessionRepository) FindByConnectionID(ctx context.Context, connectionID domain.ConnectionID) ([]*domain.Session, error) {
	var sessions []*domain.Session
	for _, session := range r.sessions {
		if session.ConnectionID == connectionID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *MemorySessionRepository) FindActive(ctx context.Context) ([]*domain.Session, error) {
	var sessions []*domain.Session
	for _, session := range r.sessions {
		if session.State == domain.SessionStateActive {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *MemorySessionRepository) Delete(ctx context.Context, sessionID domain.SessionID) error {
	delete(r.sessions, sessionID)
	return nil
}