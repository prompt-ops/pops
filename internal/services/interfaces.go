package services

import (
	"context"

	"github.com/prompt-ops/pops/internal/domain"
)

type ConnectionServiceInterface interface {
	CreateConnection(ctx context.Context, req CreateConnectionRequest) (*domain.Connection, error)
	GetConnection(ctx context.Context, name string) (*domain.Connection, error)
	ListConnections(ctx context.Context, filter ConnectionFilter) ([]*domain.Connection, error)
	DeleteConnection(ctx context.Context, name string) error
	RefreshContext(ctx context.Context, name string) error
}

type CommandServiceInterface interface {
	GenerateCommand(ctx context.Context, req GenerateCommandRequest) (*domain.Command, error)
	GenerateAnswer(ctx context.Context, req GenerateAnswerRequest) (string, error)
	ExecuteCommand(ctx context.Context, commandID domain.CommandID, dryRun bool) (*domain.ExecutionResult, error)
	GetCommandHistory(ctx context.Context, connectionName string) ([]*domain.Command, error)
}

type SessionServiceInterface interface {
	CreateSession(ctx context.Context, connectionID domain.ConnectionID) (*domain.Session, error)
	GetSession(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error)
	AddInteraction(ctx context.Context, sessionID domain.SessionID, interaction domain.Interaction) error
	UpdateSessionState(ctx context.Context, sessionID domain.SessionID, state domain.SessionState) error
	CloseSession(ctx context.Context, sessionID domain.SessionID) error
	GetActiveSessions(ctx context.Context) ([]*domain.Session, error)
}

type ShellServiceInterface interface {
	StartSession(ctx context.Context, connectionName string) (*domain.Session, error)
	ProcessInput(ctx context.Context, connectionName string, input string, mode domain.QueryMode) (*ShellResponse, error)
	ExecuteCommand(ctx context.Context, sessionID domain.SessionID, commandID domain.CommandID) (*domain.ExecutionResult, error)
	ShowContext(ctx context.Context, connectionName string) (*ContextDisplay, error)
}
