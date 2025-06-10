package domain

import (
	"context"
	"time"
)

type ConnectionID string
type ConnectionType string
type ConnectionStatus string

const (
	ConnectionTypeDatabase   ConnectionType = "database"
	ConnectionTypeKubernetes ConnectionType = "kubernetes"
	ConnectionTypeCloud      ConnectionType = "cloud"
)

const (
	ConnectionStatusActive     ConnectionStatus = "active"
	ConnectionStatusInactive   ConnectionStatus = "inactive"
	ConnectionStatusError      ConnectionStatus = "error"
	ConnectionStatusValidating ConnectionStatus = "validating"
)

type Connection struct {
	ID       ConnectionID     `json:"id"`
	Name     string           `json:"name"`
	Type     ConnectionType   `json:"type"`
	Subtype  string           `json:"subtype"`
	Config   ConnectionConfig `json:"config"`
	Context  string           `json:"context"`
	Status   ConnectionStatus `json:"status"`
	LastUsed *time.Time       `json:"last_used,omitempty"`
}

type ConnectionConfig struct {
	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Database string            `json:"database,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type ConnectionRepository interface {
	Save(ctx context.Context, conn *Connection) error
	FindByName(ctx context.Context, name string) (*Connection, error)
	FindAll(ctx context.Context) ([]*Connection, error)
	Delete(ctx context.Context, name string) error
}

type ConnectionAdapter interface {
	TestConnection(ctx context.Context, conn *Connection) error
	FetchContext(ctx context.Context, conn *Connection) (string, error)
	ExecuteCommand(ctx context.Context, conn *Connection, command string) (string, error)
}
