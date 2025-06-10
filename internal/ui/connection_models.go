package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/internal/adapters"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/repositories"
	"github.com/prompt-ops/pops/internal/services"
	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
)

// ConnectionCreateModel handles connection creation UI
type ConnectionCreateModel struct {
	// TODO: Implement simplified connection creation
	message string
}

func NewConnectionCreateModel() *ConnectionCreateModel {
	return &ConnectionCreateModel{
		message: "Connection creation not yet implemented",
	}
}

func (m *ConnectionCreateModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectionCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConnectionCreateModel) View() string {
	return fmt.Sprintf("%s\n\nPress 'q' to quit", m.message)
}

// ConnectionListModel handles listing connections
type ConnectionListModel struct {
	message string
}

func NewConnectionListModel() *ConnectionListModel {
	return &ConnectionListModel{
		message: "Connection listing not yet implemented",
	}
}

func (m *ConnectionListModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConnectionListModel) View() string {
	return fmt.Sprintf("%s\n\nPress 'q' to quit", m.message)
}

// ConnectionOpenModel handles opening connection shells
type ConnectionOpenModel struct {
	step       int
	table      table.Model
	shell      tea.Model
	err        error
}

const (
	stepSelectConnection = iota
	stepOpenShell
)

func NewConnectionOpenModel() *ConnectionOpenModel {
	// Get connections from config
	connections, err := config.GetAllConnections()
	if err != nil {
		return &ConnectionOpenModel{
			err: err,
		}
	}

	// Create table
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 15},
		{Title: "Subtype", Width: 20},
	}

	rows := make([]table.Row, len(connections))
	for i, conn := range connections {
		rows[i] = table.Row{conn.Name, conn.Type.GetMainType(), conn.Type.GetSubtype()}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("212")).
		Bold(true)
	t.SetStyles(s)

	return &ConnectionOpenModel{
		step:  stepSelectConnection,
		table: t,
	}
}

func (m *ConnectionOpenModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectionOpenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch m.step {
	case stepSelectConnection:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				// Get selected connection
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					connectionName := selectedRow[0]
					return m.openConnection(connectionName)
				}
			}
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case stepOpenShell:
		if m.shell != nil {
			var cmd tea.Cmd
			m.shell, cmd = m.shell.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *ConnectionOpenModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	switch m.step {
	case stepSelectConnection:
		return "Select a connection to open:\n\n" + m.table.View() + "\n\nPress Enter to select, Ctrl+C to quit"
	case stepOpenShell:
		if m.shell != nil {
			return m.shell.View()
		}
		return "Opening shell..."
	}

	return "Unknown state"
}

func (m *ConnectionOpenModel) openConnection(connectionName string) (*ConnectionOpenModel, tea.Cmd) {
	// Get connection from config
	conn, err := config.GetConnectionByName(connectionName)
	if err != nil {
		m.err = fmt.Errorf("failed to get connection: %w", err)
		return m, nil
	}

	// Convert to domain connection
	domainConn := ConvertToDomainConnection(conn)
	
	// Create shell service
	shellService := CreateShellService(domainConn)
	
	// Create simple shell
	shell := NewSimpleShell(connectionName, shellService)
	
	m.step = stepOpenShell
	m.shell = shell
	
	return m, shell.Init()
}

// ConnectionDeleteModel handles connection deletion
type ConnectionDeleteModel struct {
	message string
}

func NewConnectionDeleteModel() *ConnectionDeleteModel {
	return &ConnectionDeleteModel{
		message: "Connection deletion not yet implemented",
	}
}

func (m *ConnectionDeleteModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectionDeleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConnectionDeleteModel) View() string {
	return fmt.Sprintf("%s\n\nPress 'q' to quit", m.message)
}

// Helper functions (moved from the old open.go)
func ConvertToDomainConnection(oldConn conn.Connection) domain.Connection {
	var connType domain.ConnectionType
	var subtype string

	mainType := oldConn.Type.GetMainType()
	switch mainType {
	case "Kubernetes":
		connType = domain.ConnectionTypeKubernetes
		subtype = oldConn.Type.GetSubtype()
	case "Database":
		connType = domain.ConnectionTypeDatabase
		subtype = oldConn.Type.GetSubtype()
	case "Cloud":
		connType = domain.ConnectionTypeCloud
		subtype = oldConn.Type.GetSubtype()
	default:
		connType = domain.ConnectionTypeKubernetes // fallback
		subtype = "unknown"
	}

	// Try to fetch real context from the connection
	contextString := "Auto-generated context" // fallback
	if connImpl, err := conn.GetConnection(oldConn); err == nil {
		if err := connImpl.SetContext(); err == nil {
			contextString = connImpl.GetContext()
		}
	}

	return domain.Connection{
		ID:      domain.ConnectionID(oldConn.Name),
		Name:    oldConn.Name,
		Type:    connType,
		Subtype: subtype,
		Status:  domain.ConnectionStatusActive, // assume active
		Context: contextString,
	}
}

func CreateShellService(conn domain.Connection) services.ShellServiceInterface {
	// Create repositories
	connRepo := repositories.NewMemoryConnectionRepository()

	// Create adapters based on connection type
	var connectionAdapter domain.ConnectionAdapter
	switch conn.Type {
	case domain.ConnectionTypeKubernetes:
		connectionAdapter = adapters.NewFakeKubernetesAdapter()
	case domain.ConnectionTypeDatabase:
		connectionAdapter = adapters.NewFakeDatabaseAdapter()
	case domain.ConnectionTypeCloud:
		connectionAdapter = adapters.NewFakeCloudAdapter()
	default:
		connectionAdapter = adapters.NewFakeKubernetesAdapter() // fallback
	}

	// Update connection with adapter context if it's still placeholder
	if conn.Context == "Auto-generated context" {
		if adapterContext, err := connectionAdapter.FetchContext(context.Background(), &conn); err == nil {
			conn.Context = adapterContext
		}
	}

	// Add the connection to repository
	if err := connRepo.SaveValue(conn); err != nil {
		// This is unlikely to fail for an in-memory repository, but handle it anyway
		panic(fmt.Errorf("failed to save connection to repository: %w", err))
	}

	// Create services
	connService := services.NewConnectionService(connRepo, connectionAdapter)

	// Use real AI-powered command service instead of mock
	commandService := &RealAICommandService{
		connectionName: conn.Name,
	}
	sessionService := &MockSessionService{}

	return services.NewShellService(connService, commandService, sessionService)
}

// RealAICommandService uses the existing OpenAI integration
type RealAICommandService struct {
	connectionName string
}

func (r *RealAICommandService) GenerateCommand(ctx context.Context, req services.GenerateCommandRequest) (*domain.Command, error) {
	// Get the original connection from config to use existing AI integration
	oldConn, err := config.GetConnectionByName(r.connectionName)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	// Get the proper connection implementation with AI integration
	connImpl, err := conn.GetConnection(oldConn)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection implementation: %w", err)
	}

	// Set context (this populates the AI context with actual database/cluster info)
	if err := connImpl.SetContext(); err != nil {
		return nil, fmt.Errorf("failed to set connection context: %w", err)
	}

	// Use the real AI-powered command generation
	commandText, err := connImpl.GetCommand(req.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	// Create domain command with proper safety assessment
	command := &domain.Command{
		ID:              domain.CommandID(fmt.Sprintf("cmd-%d", time.Now().Unix())),
		Type:            domain.CommandTypeQuery,
		Content:         commandText,
		Description:     fmt.Sprintf("Generated command for: %s", req.Prompt),
		UserPrompt:      req.Prompt,
		Safety:          determineSafetyLevel(commandText),
		EstimatedImpact: assessCommandImpact(commandText),
		GeneratedAt:     time.Now(),
		Status:          domain.ExecutionStatusPending,
	}

	return command, nil
}

func (r *RealAICommandService) GenerateAnswer(ctx context.Context, req services.GenerateAnswerRequest) (string, error) {
	// Get the original connection from config
	oldConn, err := config.GetConnectionByName(r.connectionName)
	if err != nil {
		return "", fmt.Errorf("connection not found: %w", err)
	}

	// Get the proper connection implementation
	connImpl, err := conn.GetConnection(oldConn)
	if err != nil {
		return "", fmt.Errorf("failed to get connection implementation: %w", err)
	}

	// Set context
	if err := connImpl.SetContext(); err != nil {
		return "", fmt.Errorf("failed to set connection context: %w", err)
	}

	// Use the real AI-powered answer generation
	answer, err := connImpl.GetAnswer(req.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, nil
}

func (r *RealAICommandService) ExecuteCommand(ctx context.Context, commandID domain.CommandID, dryRun bool) (*domain.ExecutionResult, error) {
	// TODO: Implement real command execution
	return &domain.ExecutionResult{
		CommandID: commandID,
		Status:    domain.ExecutionStatusCompleted,
		Output:    "Command execution not yet implemented",
		Duration:  time.Second,
		ExecutedAt: time.Now(),
		ExitCode:   0,
	}, nil
}

func (r *RealAICommandService) GetCommandHistory(ctx context.Context, connectionName string) ([]*domain.Command, error) {
	// TODO: Implement command history retrieval
	return []*domain.Command{}, nil
}

// Helper functions for safety assessment
func determineSafetyLevel(command string) domain.SafetyLevel {
	command = strings.ToLower(command)
	
	if strings.Contains(command, "delete") || strings.Contains(command, "drop") || 
	   strings.Contains(command, "truncate") || strings.Contains(command, "remove") {
		return domain.SafetyLevelDestructive
	}
	
	if strings.Contains(command, "update") || strings.Contains(command, "insert") || 
	   strings.Contains(command, "create") || strings.Contains(command, "alter") {
		return domain.SafetyLevelCaution
	}
	
	return domain.SafetyLevelSafe
}

func assessCommandImpact(command string) domain.ImpactAssessment {
	safety := determineSafetyLevel(command)
	
	impact := domain.ImpactAssessment{
		RiskLevel:        safety,
		RequiresApproval: safety == domain.SafetyLevelDestructive,
		Metadata:         make(map[string]string),
	}
	
	if safety == domain.SafetyLevelDestructive {
		impact.Warnings = []string{
			"This command may delete or modify data permanently",
			"Please review carefully before executing",
		}
	}
	
	return impact
}

// Keep MockCommandService for backward compatibility
type MockCommandService struct {}

func (m *MockCommandService) GenerateCommand(ctx context.Context, req services.GenerateCommandRequest) (*domain.Command, error) {
	return &domain.Command{
		ID:          "cmd-1",
		Type:        domain.CommandTypeQuery,
		Content:     "mock command",
		Description: "Mock command for testing",
		Safety:      domain.SafetyLevelSafe,
	}, nil
}

func (m *MockCommandService) GenerateAnswer(ctx context.Context, req services.GenerateAnswerRequest) (string, error) {
	return "This is a mock answer.", nil
}

func (m *MockCommandService) ExecuteCommand(ctx context.Context, commandID domain.CommandID, dryRun bool) (*domain.ExecutionResult, error) {
	return &domain.ExecutionResult{
		CommandID: commandID,
		Status:    domain.ExecutionStatusCompleted,
		Output:    "Mock command executed successfully",
	}, nil
}

func (m *MockCommandService) GetCommandHistory(ctx context.Context, connectionName string) ([]*domain.Command, error) {
	return []*domain.Command{}, nil
}

// MockSessionService for temporary use
type MockSessionService struct{}

func (m *MockSessionService) CreateSession(ctx context.Context, connectionID domain.ConnectionID) (*domain.Session, error) {
	return &domain.Session{
		ID:           "session-1",
		ConnectionID: connectionID,
		State:        domain.SessionStateActive,
	}, nil
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error) {
	return &domain.Session{
		ID:    sessionID,
		State: domain.SessionStateActive,
	}, nil
}

func (m *MockSessionService) AddInteraction(ctx context.Context, sessionID domain.SessionID, interaction domain.Interaction) error {
	return nil
}

func (m *MockSessionService) UpdateSessionState(ctx context.Context, sessionID domain.SessionID, state domain.SessionState) error {
	return nil
}

func (m *MockSessionService) CloseSession(ctx context.Context, sessionID domain.SessionID) error {
	return nil
}

func (m *MockSessionService) GetActiveSessions(ctx context.Context) ([]*domain.Session, error) {
	return []*domain.Session{}, nil
}