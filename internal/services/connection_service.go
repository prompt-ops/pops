package services

import (
	"context"
	"fmt"
	"time"

	"github.com/prompt-ops/pops/internal/domain"
)

type ConnectionService struct {
	repo    domain.ConnectionRepository
	adapter domain.ConnectionAdapter
}

func NewConnectionService(
	repo domain.ConnectionRepository,
	adapter domain.ConnectionAdapter,
) *ConnectionService {
	return &ConnectionService{
		repo:    repo,
		adapter: adapter,
	}
}

func (s *ConnectionService) CreateConnection(ctx context.Context, req CreateConnectionRequest) (*domain.Connection, error) {
	conn := &domain.Connection{
		ID:      domain.ConnectionID(generateID()),
		Name:    req.Name,
		Type:    req.Type,
		Subtype: req.Subtype,
		Config: domain.ConnectionConfig{
			Host:     req.Host,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
			Database: req.Database,
			Extra:    req.Metadata,
		},
		Status: domain.ConnectionStatusValidating,
	}

	if err := s.adapter.TestConnection(ctx, conn); err != nil {
		conn.Status = domain.ConnectionStatusError
		if saveErr := s.repo.Save(ctx, conn); saveErr != nil {
			return nil, fmt.Errorf("failed to save connection with error status: %w", saveErr)
		}
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	context, err := s.adapter.FetchContext(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch connection context: %w", err)
	}

	conn.Context = context
	conn.Status = domain.ConnectionStatusActive

	if err := s.repo.Save(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return conn, nil
}

func (s *ConnectionService) GetConnection(ctx context.Context, name string) (*domain.Connection, error) {
	conn, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	// Update last used time
	now := time.Now()
	conn.LastUsed = &now
	if err := s.repo.Save(ctx, conn); err != nil {
		// Log the error but don't fail the operation since this is just updating the last used time
		// In a real implementation, you might want to log this error
		// For now, we'll just ignore it to maintain backward compatibility
		_ = err
	}

	return conn, nil
}

func (s *ConnectionService) ListConnections(ctx context.Context, filter ConnectionFilter) ([]*domain.Connection, error) {
	connections, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if filter.Type != "" {
		var filtered []*domain.Connection
		for _, conn := range connections {
			if conn.Type == filter.Type {
				filtered = append(filtered, conn)
			}
		}
		return filtered, nil
	}

	return connections, nil
}

func (s *ConnectionService) DeleteConnection(ctx context.Context, name string) error {
	return s.repo.Delete(ctx, name)
}

func (s *ConnectionService) RefreshContext(ctx context.Context, name string) error {
	conn, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}

	context, err := s.adapter.FetchContext(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to fetch context: %w", err)
	}

	conn.Context = context
	return s.repo.Save(ctx, conn)
}

type CreateConnectionRequest struct {
	Name     string                `json:"name"`
	Type     domain.ConnectionType `json:"type"`
	Subtype  string                `json:"subtype"`
	Host     string                `json:"host"`
	Port     int                   `json:"port"`
	Username string                `json:"username"`
	Password string                `json:"password"`
	Database string                `json:"database"`
	Metadata map[string]string     `json:"metadata"`
}

type ConnectionFilter struct {
	Type domain.ConnectionType `json:"type"`
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
