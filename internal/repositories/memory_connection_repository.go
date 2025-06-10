package repositories

import (
	"context"
	"fmt"

	"github.com/prompt-ops/pops/internal/domain"
)

type MemoryConnectionRepository struct {
	connections map[string]*domain.Connection
}

func NewMemoryConnectionRepository() *MemoryConnectionRepository {
	return &MemoryConnectionRepository{
		connections: make(map[string]*domain.Connection),
	}
}

func (r *MemoryConnectionRepository) Save(ctx context.Context, conn *domain.Connection) error {
	r.connections[conn.Name] = conn
	return nil
}

func (r *MemoryConnectionRepository) SaveValue(conn domain.Connection) error {
	r.connections[conn.Name] = &conn
	return nil
}

func (r *MemoryConnectionRepository) FindByName(ctx context.Context, name string) (*domain.Connection, error) {
	conn, exists := r.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", name)
	}
	return conn, nil
}

func (r *MemoryConnectionRepository) FindAll(ctx context.Context) ([]*domain.Connection, error) {
	var connections []*domain.Connection
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *MemoryConnectionRepository) Delete(ctx context.Context, name string) error {
	delete(r.connections, name)
	return nil
}