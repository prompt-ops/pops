package repositories

import (
	"context"
	"fmt"

	"github.com/prompt-ops/pops/internal/domain"
)

type MemoryCommandRepository struct {
	commands map[domain.CommandID]*domain.Command
}

func NewMemoryCommandRepository() *MemoryCommandRepository {
	return &MemoryCommandRepository{
		commands: make(map[domain.CommandID]*domain.Command),
	}
}

func (r *MemoryCommandRepository) Save(ctx context.Context, command *domain.Command) error {
	r.commands[command.ID] = command
	return nil
}

func (r *MemoryCommandRepository) FindByID(ctx context.Context, commandID domain.CommandID) (*domain.Command, error) {
	command, exists := r.commands[commandID]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandID)
	}
	return command, nil
}

func (r *MemoryCommandRepository) FindByConnection(ctx context.Context, connectionID domain.ConnectionID) ([]*domain.Command, error) {
	var commands []*domain.Command
	for _, command := range r.commands {
		if command.ConnectionID == connectionID {
			commands = append(commands, command)
		}
	}
	return commands, nil
}

func (r *MemoryCommandRepository) SaveResult(ctx context.Context, result *domain.ExecutionResult) error {
	// In a real implementation, this would save execution results
	return nil
}

func (r *MemoryCommandRepository) FindResult(ctx context.Context, commandID domain.CommandID) (*domain.ExecutionResult, error) {
	// In a real implementation, this would find execution results
	return nil, fmt.Errorf("result not found for command: %s", commandID)
}

func (r *MemoryCommandRepository) Delete(ctx context.Context, commandID domain.CommandID) error {
	delete(r.commands, commandID)
	return nil
}