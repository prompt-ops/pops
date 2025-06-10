package shell

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
)

// SimpleShell is a clean, focused shell implementation
type SimpleShell struct {
	// Core
	connectionName string
	shellService   services.ShellServiceInterface
	
	// UI Components
	input textinput.Model
	
	// State
	state   State
	mode    domain.QueryMode
	
	// Data
	context  *services.ContextDisplay
	response *services.ShellResponse
	history  []HistoryEntry
	
	// Layout
	width  int
	height int
	
	// Error
	err error
}

type State int

const (
	StateInput State = iota
	StateShowingContext
	StateConfirmingCommand
	StateShowingResult
	StateError
)

type HistoryEntry struct {
	Prompt   string
	Response string
	Mode     domain.QueryMode
	Success  bool
}

// NewSimpleShell creates a new simplified shell
func NewSimpleShell(connectionName string, shellService services.ShellServiceInterface) *SimpleShell {
	input := textinput.New()
	input.Placeholder = "What would you like to do? (Tab: switch mode, F1: show context)"
	input.Focus()
	input.CharLimit = 500

	return &SimpleShell{
		connectionName: connectionName,
		shellService:   shellService,
		input:          input,
		state:          StateInput,
		mode:           domain.QueryModeCommand,
		history:        []HistoryEntry{},
	}
}

func (m *SimpleShell) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SimpleShell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = min(m.width-10, 80)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case contextLoadedMsg:
		m.context = msg.context
		m.state = StateShowingContext
		return m, nil

	case responseGeneratedMsg:
		m.response = msg.response
		if m.response.Type == services.ResponseTypeCommand {
			m.state = StateConfirmingCommand
		} else {
			m.addToHistory(m.input.Value(), m.response.Answer, true)
			m.resetToInput()
		}
		return m, nil

	case resultMsg:
		m.addToHistory(m.input.Value(), msg.output, msg.success)
		m.resetToInput()
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.state = StateError
		return m, nil
	}

	// Update input when in input state
	if m.state == StateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SimpleShell) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateInput:
		return m.handleInputKeys(msg)
	case StateShowingContext, StateError:
		return m.handleViewKeys(msg)
	case StateConfirmingCommand:
		return m.handleConfirmKeys(msg)
	case StateShowingResult:
		return m.handleResultKeys(msg)
	default:
		return m, nil
	}
}

func (m *SimpleShell) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyF1:
		return m, m.loadContext()

	case tea.KeyTab:
		m.toggleMode()
		return m, nil

	case tea.KeyEnter:
		prompt := strings.TrimSpace(m.input.Value())
		if prompt != "" {
			return m, m.processInput(prompt)
		}

	case tea.KeyUp:
		m.navigateHistory(-1)
		return m, nil

	case tea.KeyDown:
		m.navigateHistory(1)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *SimpleShell) handleViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter, tea.KeyEsc:
		m.resetToInput()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *SimpleShell) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.response != nil {
			return m, m.executeCommand(m.response.Command)
		}
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeyEsc:
		m.resetToInput()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *SimpleShell) handleResultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter, tea.KeyEsc:
		m.resetToInput()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *SimpleShell) View() string {
	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader())
	content.WriteString("\n\n")

	// Main content
	switch m.state {
	case StateInput:
		content.WriteString(m.renderInput())
	case StateShowingContext:
		content.WriteString(m.renderContext())
	case StateConfirmingCommand:
		content.WriteString(m.renderConfirmation())
	case StateShowingResult:
		content.WriteString(m.renderResult())
	case StateError:
		content.WriteString(m.renderError())
	}

	// History (if space allows)
	if m.height > 15 && len(m.history) > 0 {
		content.WriteString("\n\n")
		content.WriteString(m.renderHistory())
	}

	return content.String()
}

func (m *SimpleShell) renderHeader() string {
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("63")).
		Padding(0, 1).
		Bold(true)

	connectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	modeText := "COMMAND"
	if m.mode == domain.QueryModeAnswer {
		modeText = "QUESTION"
	}

	return fmt.Sprintf("%s %s",
		modeStyle.Render(modeText),
		connectionStyle.Render(fmt.Sprintf("→ %s", m.connectionName)),
	)
}

func (m *SimpleShell) renderInput() string {
	var content strings.Builder
	
	content.WriteString(m.input.View())
	content.WriteString("\n\n")
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString(helpStyle.Render("Tab: switch mode • F1: show context • ↑↓: history • Enter: submit"))

	return content.String()
}

func (m *SimpleShell) renderContext() string {
	if m.context == nil {
		return "No context available\n\nPress Enter to return"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render(m.context.Title))
	content.WriteString("\n")
	content.WriteString(m.context.Description)
	content.WriteString("\n\n")

	// Group items by type for cleaner display
	groups := make(map[string][]services.ContextItem)
	for _, item := range m.context.Items {
		groups[item.Type] = append(groups[item.Type], item)
	}

	for groupType, items := range groups {
		if len(items) == 0 {
			continue
		}

		groupStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

		content.WriteString(groupStyle.Render(fmt.Sprintf("%s (%d):", strings.ToUpper(groupType[:1])+groupType[1:], len(items))))
		content.WriteString("\n")

		for i, item := range items {
			if i >= 5 {
				content.WriteString(fmt.Sprintf("  ... and %d more\n", len(items)-5))
				break
			}
			content.WriteString(fmt.Sprintf("  %s %s\n", item.Icon, item.Name))
		}
		content.WriteString("\n")
	}

	content.WriteString("\nPress Enter to return")
	return content.String()
}

func (m *SimpleShell) renderConfirmation() string {
	if m.response == nil {
		return "No command to confirm"
	}

	commandStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	safetyStyle := m.getSafetyStyle(m.response.Safety)

	var content strings.Builder
	content.WriteString("Generated command:\n\n")
	content.WriteString(commandStyle.Render(m.response.Command))
	content.WriteString("\n\n")
	content.WriteString(safetyStyle.Render(fmt.Sprintf("Safety Level: %s", m.response.Safety)))

	if len(m.response.Impact.Warnings) > 0 {
		content.WriteString("\n\nWarnings:\n")
		for _, warning := range m.response.Impact.Warnings {
			content.WriteString(fmt.Sprintf("⚠️  %s\n", warning))
		}
	}

	content.WriteString("\n\nPress Enter to execute • Esc to cancel")
	return content.String()
}

func (m *SimpleShell) renderResult() string {
	return "Result would be shown here\n\nPress Enter to continue"
}

func (m *SimpleShell) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	return fmt.Sprintf("%s\n\nPress Enter to continue", 
		errorStyle.Render(fmt.Sprintf("Error: %s", m.err)))
}

func (m *SimpleShell) renderHistory() string {
	if len(m.history) == 0 {
		return ""
	}

	historyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("236")).
		Padding(0, 0, 0, 2)

	var content strings.Builder
	content.WriteString("Recent:\n")

	// Show last 3 items
	start := max(0, len(m.history)-3)
	for i := start; i < len(m.history); i++ {
		item := m.history[i]
		icon := "✓"
		if !item.Success {
			icon = "✗"
		}
		modeIcon := "🔧"
		if item.Mode == domain.QueryModeAnswer {
			modeIcon = "❓"
		}
		content.WriteString(fmt.Sprintf("%s %s %s\n", icon, modeIcon, truncate(item.Prompt, 40)))
	}

	return historyStyle.Render(content.String())
}

// Helper methods
func (m *SimpleShell) toggleMode() {
	if m.mode == domain.QueryModeCommand {
		m.mode = domain.QueryModeAnswer
		m.input.Placeholder = "Ask a question..."
	} else {
		m.mode = domain.QueryModeCommand
		m.input.Placeholder = "What would you like to do?"
	}
}

func (m *SimpleShell) resetToInput() {
	m.state = StateInput
	m.input.Reset()
	m.err = nil
	m.response = nil
}

func (m *SimpleShell) addToHistory(prompt, response string, success bool) {
	m.history = append(m.history, HistoryEntry{
		Prompt:   prompt,
		Response: response,
		Mode:     m.mode,
		Success:  success,
	})

	// Keep only last 10 items
	if len(m.history) > 10 {
		m.history = m.history[1:]
	}
}

func (m *SimpleShell) navigateHistory(direction int) {
	if len(m.history) == 0 {
		return
	}

	// Simple history navigation - just get the last few prompts
	if direction < 0 && len(m.history) > 0 {
		lastEntry := m.history[len(m.history)-1]
		m.input.SetValue(lastEntry.Prompt)
		m.input.CursorEnd()
	}
}

func (m *SimpleShell) getSafetyStyle(safety domain.SafetyLevel) lipgloss.Style {
	switch safety {
	case domain.SafetyLevelSafe:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	case domain.SafetyLevelCaution:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	case domain.SafetyLevelDangerous:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case domain.SafetyLevelDestructive:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	default:
		return lipgloss.NewStyle()
	}
}

// Commands
func (m *SimpleShell) loadContext() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		display, err := m.shellService.ShowContext(ctx, m.connectionName)
		if err != nil {
			return errorMsg{err: err}
		}
		return contextLoadedMsg{context: display}
	}
}

func (m *SimpleShell) processInput(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.shellService.ProcessInput(ctx, m.connectionName, prompt, m.mode)
		if err != nil {
			return errorMsg{err: err}
		}
		return responseGeneratedMsg{response: response}
	}
}

func (m *SimpleShell) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		// Simulate command execution
		return resultMsg{
			output:  fmt.Sprintf("Executed: %s\nResult: Command completed successfully", command),
			success: true,
		}
	}
}

// Messages
type contextLoadedMsg struct {
	context *services.ContextDisplay
}

type responseGeneratedMsg struct {
	response *services.ShellResponse
}

type resultMsg struct {
	output  string
	success bool
}

type errorMsg struct {
	err error
}

// Utility functions
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}