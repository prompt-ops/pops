package domain

import (
	"context"
	"time"
)

type SessionID string
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStateInactive  SessionState = "inactive"
	SessionStateCompleted SessionState = "completed"
)

type Session struct {
	ID           SessionID     `json:"id"`
	ConnectionID ConnectionID  `json:"connection_id"`
	State        SessionState  `json:"state"`
	History      []Interaction `json:"history"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	LastActive   time.Time     `json:"last_active"`
}

type Interaction struct {
	ID        string    `json:"id"`
	Prompt    string    `json:"prompt"`
	Command   string    `json:"command,omitempty"`
	Answer    string    `json:"answer,omitempty"`
	Mode      QueryMode `json:"mode"`
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type SessionService interface {
	CreateSession(ctx context.Context, connectionID ConnectionID) (*Session, error)
	GetSession(ctx context.Context, sessionID SessionID) (*Session, error)
	AddInteraction(ctx context.Context, sessionID SessionID, interaction Interaction) error
	UpdateSessionState(ctx context.Context, sessionID SessionID, state SessionState) error
	CloseSession(ctx context.Context, sessionID SessionID) error
	GetActiveSessions(ctx context.Context) ([]*Session, error)
}

type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id SessionID) (*Session, error)
	FindByConnection(ctx context.Context, connID ConnectionID) ([]*Session, error)
	FindActive(ctx context.Context) ([]*Session, error)
	Delete(ctx context.Context, id SessionID) error
}
