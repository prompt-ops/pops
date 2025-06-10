package domain

import (
	"context"
	"time"
)

type CommandID string
type CommandType string
type SafetyLevel string
type ExecutionStatus string

const (
	CommandTypeQuery      CommandType = "query"
	CommandTypeOperation  CommandType = "operation"
	CommandTypeDeployment CommandType = "deployment"
	CommandTypeAdmin      CommandType = "admin"
)

const (
	SafetyLevelSafe        SafetyLevel = "safe"
	SafetyLevelCaution     SafetyLevel = "caution"
	SafetyLevelDangerous   SafetyLevel = "dangerous"
	SafetyLevelDestructive SafetyLevel = "destructive"
)

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

type Command struct {
	ID              CommandID        `json:"id"`
	ConnectionID    ConnectionID     `json:"connection_id"`
	Type            CommandType      `json:"type"`
	Content         string           `json:"content"`
	Description     string           `json:"description"`
	UserPrompt      string           `json:"user_prompt"`
	Safety          SafetyLevel      `json:"safety"`
	EstimatedImpact ImpactAssessment `json:"estimated_impact"`
	GeneratedAt     time.Time        `json:"generated_at"`
	ExecutedAt      *time.Time       `json:"executed_at,omitempty"`
	Status          ExecutionStatus  `json:"status"`
}

type ImpactAssessment struct {
	AffectedResources []string          `json:"affected_resources"`
	RiskLevel         SafetyLevel       `json:"risk_level"`
	RequiresApproval  bool              `json:"requires_approval"`
	Warnings          []string          `json:"warnings"`
	Metadata          map[string]string `json:"metadata"`
}

type ExecutionResult struct {
	CommandID  CommandID       `json:"command_id"`
	Status     ExecutionStatus `json:"status"`
	Output     string          `json:"output"`
	Error      string          `json:"error,omitempty"`
	Duration   time.Duration   `json:"duration"`
	ExecutedAt time.Time       `json:"executed_at"`
	ExitCode   int             `json:"exit_code"`
}

type CommandRequest struct {
	ConnectionID ConnectionID `json:"connection_id"`
	Prompt       string       `json:"prompt"`
	Context      string       `json:"context"`
	Mode         QueryMode    `json:"mode"`
}

type QueryMode string

const (
	QueryModeCommand QueryMode = "command"
	QueryModeAnswer  QueryMode = "answer"
)

type CommandGenerator interface {
	GenerateCommand(ctx context.Context, req CommandRequest) (*Command, error)
	GenerateAnswer(ctx context.Context, req CommandRequest) (string, error)
	ValidateCommand(ctx context.Context, cmd *Command) error
}

type CommandExecutor interface {
	Execute(ctx context.Context, cmd *Command) (*ExecutionResult, error)
	DryRun(ctx context.Context, cmd *Command) (*ExecutionResult, error)
	Cancel(ctx context.Context, commandID CommandID) error
}

type CommandRepository interface {
	Save(ctx context.Context, cmd *Command) error
	FindByID(ctx context.Context, id CommandID) (*Command, error)
	FindByConnection(ctx context.Context, connID ConnectionID) ([]*Command, error)
	SaveResult(ctx context.Context, result *ExecutionResult) error
	FindResult(ctx context.Context, commandID CommandID) (*ExecutionResult, error)
}
