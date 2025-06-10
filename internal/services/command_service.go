package services

import (
	"context"
	"fmt"

	"github.com/prompt-ops/pops/internal/domain"
)

type CommandService struct {
	generator domain.CommandGenerator
	executor  domain.CommandExecutor
	cmdRepo   domain.CommandRepository
	connRepo  domain.ConnectionRepository
}

func NewCommandService(
	generator domain.CommandGenerator,
	executor domain.CommandExecutor,
	cmdRepo domain.CommandRepository,
	connRepo domain.ConnectionRepository,
) *CommandService {
	return &CommandService{
		generator: generator,
		executor:  executor,
		cmdRepo:   cmdRepo,
		connRepo:  connRepo,
	}
}

func (s *CommandService) GenerateCommand(ctx context.Context, req GenerateCommandRequest) (*domain.Command, error) {
	conn, err := s.connRepo.FindByName(ctx, req.ConnectionName)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	cmdReq := domain.CommandRequest{
		ConnectionID: conn.ID,
		Prompt:       req.Prompt,
		Context:      s.buildContext(conn),
		Mode:         domain.QueryModeCommand,
	}

	cmd, err := s.generator.GenerateCommand(ctx, cmdReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	if err := s.cmdRepo.Save(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to save command: %w", err)
	}

	return cmd, nil
}

func (s *CommandService) GenerateAnswer(ctx context.Context, req GenerateAnswerRequest) (string, error) {
	conn, err := s.connRepo.FindByName(ctx, req.ConnectionName)
	if err != nil {
		return "", fmt.Errorf("connection not found: %w", err)
	}

	cmdReq := domain.CommandRequest{
		ConnectionID: conn.ID,
		Prompt:       req.Prompt,
		Context:      s.buildContext(conn),
		Mode:         domain.QueryModeAnswer,
	}

	return s.generator.GenerateAnswer(ctx, cmdReq)
}

func (s *CommandService) ExecuteCommand(ctx context.Context, commandID domain.CommandID, dryRun bool) (*domain.ExecutionResult, error) {
	cmd, err := s.cmdRepo.FindByID(ctx, commandID)
	if err != nil {
		return nil, fmt.Errorf("command not found: %w", err)
	}

	if err := s.generator.ValidateCommand(ctx, cmd); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	var result *domain.ExecutionResult
	if dryRun {
		result, err = s.executor.DryRun(ctx, cmd)
	} else {
		result, err = s.executor.Execute(ctx, cmd)
		if err == nil {
			cmd.ExecutedAt = &result.ExecutedAt
			cmd.Status = result.Status
			if saveErr := s.cmdRepo.Save(ctx, cmd); saveErr != nil {
				return nil, fmt.Errorf("failed to update command status: %w", saveErr)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	if err := s.cmdRepo.SaveResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save result: %w", err)
	}

	return result, nil
}

func (s *CommandService) GetCommandHistory(ctx context.Context, connectionName string) ([]*domain.Command, error) {
	conn, err := s.connRepo.FindByName(ctx, connectionName)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	return s.cmdRepo.FindByConnection(ctx, conn.ID)
}

func (s *CommandService) buildContext(conn *domain.Connection) string {
	return fmt.Sprintf("Connection: %s (%s)\n%s", conn.Name, conn.Type, conn.Context)
}

type GenerateCommandRequest struct {
	ConnectionName string `json:"connection_name"`
	Prompt         string `json:"prompt"`
}

type GenerateAnswerRequest struct {
	ConnectionName string `json:"connection_name"`
	Prompt         string `json:"prompt"`
}
