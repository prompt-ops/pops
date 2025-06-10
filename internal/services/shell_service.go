package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/internal/domain"
)

type ShellService struct {
	connectionService ConnectionServiceInterface
	commandService    CommandServiceInterface
	sessionService    SessionServiceInterface
}

func NewShellService(
	connectionService ConnectionServiceInterface,
	commandService CommandServiceInterface,
	sessionService SessionServiceInterface,
) *ShellService {
	return &ShellService{
		connectionService: connectionService,
		commandService:    commandService,
		sessionService:    sessionService,
	}
}

func (s *ShellService) StartSession(ctx context.Context, connectionName string) (*domain.Session, error) {
	conn, err := s.connectionService.GetConnection(ctx, connectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	session, err := s.sessionService.CreateSession(ctx, conn.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *ShellService) ProcessInput(ctx context.Context, connectionName string, input string, mode domain.QueryMode) (*ShellResponse, error) {
	conn, err := s.connectionService.GetConnection(ctx, connectionName)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	var response *ShellResponse

	switch mode {
	case domain.QueryModeCommand:
		cmd, err := s.commandService.GenerateCommand(ctx, GenerateCommandRequest{
			ConnectionName: conn.Name,
			Prompt:         input,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate command: %w", err)
		}

		response = &ShellResponse{
			Type:    ResponseTypeCommand,
			Command: cmd.Content,
			Safety:  cmd.Safety,
			Impact:  cmd.EstimatedImpact,
		}

	case domain.QueryModeAnswer:
		answer, err := s.commandService.GenerateAnswer(ctx, GenerateAnswerRequest{
			ConnectionName: conn.Name,
			Prompt:         input,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate answer: %w", err)
		}

		response = &ShellResponse{
			Type:   ResponseTypeAnswer,
			Answer: answer,
		}
	}

	// Note: In a real implementation, you'd save this interaction to the session

	return response, nil
}

func (s *ShellService) ExecuteCommand(ctx context.Context, sessionID domain.SessionID, commandID domain.CommandID) (*domain.ExecutionResult, error) {
	result, err := s.commandService.ExecuteCommand(ctx, commandID, false)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	session, _ := s.sessionService.GetSession(ctx, sessionID)
	if session != nil && len(session.History) > 0 {
		lastInteraction := &session.History[len(session.History)-1]
		lastInteraction.Output = result.Output
		if result.Error != "" {
			lastInteraction.Error = result.Error
		}
		if err := s.sessionService.AddInteraction(ctx, sessionID, *lastInteraction); err != nil {
			// Log the error but don't fail the operation since the command already executed
			// In a real implementation, you might want to log this error
			// For now, we'll just ignore it to maintain backward compatibility
			_ = err
		}
	}

	return result, nil
}

func (s *ShellService) ShowContext(ctx context.Context, connectionName string) (*ContextDisplay, error) {
	conn, err := s.connectionService.GetConnection(ctx, connectionName)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	display := &ContextDisplay{
		ConnectionName: conn.Name,
		ConnectionType: string(conn.Type),
		Subtype:        conn.Subtype,
		Status:         string(conn.Status),
	}

	switch conn.Type {
	case domain.ConnectionTypeDatabase:
		display.Items = s.formatDatabaseContext(conn.Context)
		display.Title = "📊 Database Schema"
		display.Description = "Tables and columns available in this database"

	case domain.ConnectionTypeKubernetes:
		display.Items = s.formatKubernetesContext(conn.Context)
		display.Title = "⚙️ Kubernetes Resources"
		display.Description = "Available resources in this cluster"

	case domain.ConnectionTypeCloud:
		display.Items = s.formatCloudContext(conn.Context)
		display.Title = "☁️ Cloud Resources"
		display.Description = "Available resources in this cloud environment"
	}

	return display, nil
}

func (s *ShellService) formatDatabaseContext(context string) []ContextItem {
	var items []ContextItem
	lines := strings.Split(context, "\n")

	var currentTable string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Handle fake adapter format: "Table: tablename"
		if strings.HasPrefix(line, "Table:") {
			currentTable = strings.TrimPrefix(line, "Table: ")
			items = append(items, ContextItem{
				Icon:        "📋",
				Name:        currentTable,
				Type:        "table",
				Description: "Database table",
			})
		} else if strings.HasPrefix(line, "  - ") && currentTable != "" {
			column := strings.TrimPrefix(line, "  - ")
			items = append(items, ContextItem{
				Icon:        "  📝",
				Name:        column,
				Type:        "column",
				Parent:      currentTable,
				Description: "Table column",
			})
		}
		// Handle real PostgreSQL format: "- **"schema"."table"**:"
		if strings.HasPrefix(line, "- **") && strings.HasSuffix(line, "**:") {
			// Extract table name from "- **"schema"."table"**:"
			tablePart := strings.TrimPrefix(line, "- **")
			tablePart = strings.TrimSuffix(tablePart, "**:")
			currentTable = tablePart
			items = append(items, ContextItem{
				Icon:        "📋",
				Name:        currentTable,
				Type:        "table",
				Description: "Database table",
			})
		} else if strings.HasPrefix(line, "  - `") && strings.Contains(line, "` (") && currentTable != "" {
			// Extract column from "  - `"column"` (type)"
			columnPart := strings.TrimPrefix(line, "  - `")
			if idx := strings.Index(columnPart, "` ("); idx > 0 {
				column := columnPart[:idx]
				typeInfo := columnPart[idx+3:]
				if idx := strings.Index(typeInfo, ")"); idx > 0 {
					typeInfo = typeInfo[:idx]
				}
				items = append(items, ContextItem{
					Icon:        "  📝",
					Name:        fmt.Sprintf("%s (%s)", column, typeInfo),
					Type:        "column",
					Parent:      currentTable,
					Description: "Table column",
				})
			}
		}
	}

	return items
}

func (s *ShellService) formatKubernetesContext(context string) []ContextItem {
	var items []ContextItem
	lines := strings.Split(context, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				resourceType := strings.TrimSpace(parts[0])
				resourceName := strings.TrimSpace(parts[1])

				icon := s.getKubernetesIcon(resourceType)
				items = append(items, ContextItem{
					Icon:        icon,
					Name:        resourceName,
					Type:        resourceType,
					Description: fmt.Sprintf("Kubernetes %s", resourceType),
				})
			}
		}
	}

	return items
}

func (s *ShellService) formatCloudContext(context string) []ContextItem {
	var items []ContextItem
	lines := strings.Split(context, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				resourceType := strings.TrimSpace(parts[0])
				resourceName := strings.TrimSpace(parts[1])

				icon := s.getCloudIcon(resourceType)
				items = append(items, ContextItem{
					Icon:        icon,
					Name:        resourceName,
					Type:        resourceType,
					Description: fmt.Sprintf("Cloud %s", resourceType),
				})
			}
		}
	}

	return items
}

func (s *ShellService) getKubernetesIcon(resourceType string) string {
	icons := map[string]string{
		"pod":        "🟢",
		"deployment": "🚀",
		"service":    "🌐",
		"configmap":  "📄",
		"secret":     "🔐",
		"ingress":    "🌍",
		"namespace":  "📁",
		"node":       "🖥️",
		"volume":     "💾",
	}

	if icon, exists := icons[strings.ToLower(resourceType)]; exists {
		return icon
	}
	return "⚙️"
}

func (s *ShellService) getCloudIcon(resourceType string) string {
	icons := map[string]string{
		"vm":           "🖥️",
		"storage":      "💾",
		"network":      "🌐",
		"loadbalancer": "⚖️",
		"database":     "🗄️",
		"container":    "📦",
		"function":     "⚡",
		"queue":        "📬",
	}

	if icon, exists := icons[strings.ToLower(resourceType)]; exists {
		return icon
	}
	return "☁️"
}

type ResponseType string

const (
	ResponseTypeCommand ResponseType = "command"
	ResponseTypeAnswer  ResponseType = "answer"
)

type ShellResponse struct {
	Type    ResponseType            `json:"type"`
	Command string                  `json:"command,omitempty"`
	Answer  string                  `json:"answer,omitempty"`
	Safety  domain.SafetyLevel      `json:"safety,omitempty"`
	Impact  domain.ImpactAssessment `json:"impact,omitempty"`
}

type ContextDisplay struct {
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	ConnectionName string        `json:"connection_name"`
	ConnectionType string        `json:"connection_type"`
	Subtype        string        `json:"subtype"`
	Status         string        `json:"status"`
	Items          []ContextItem `json:"items"`
}

type ContextItem struct {
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Parent      string `json:"parent,omitempty"`
	Description string `json:"description"`
}
