package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
)

// SimpleShell is a clean, focused shell implementation
type SimpleShell struct {
	// Core
	connectionName string
	shellService   services.ShellServiceInterface
	
	// UI Components
	input   textinput.Model
	spinner spinner.Model
	
	// State
	state   ShellState
	mode    domain.QueryMode
	
	// Data
	context       *services.ContextDisplay
	response      *services.ShellResponse
	result        *ExecutionResult
	history       []HistoryEntry
	loadingPrompt string
	
	// Scrolling for results and command confirmation
	scrollOffset        int
	maxScroll           int
	verticalOffset      int
	maxVertical         int
	confirmScrollOffset int
	confirmMaxScroll    int
	confirmVertOffset   int
	confirmMaxVertical  int
	
	// History navigation
	historyIndex        int
	selectedHistory     *HistoryEntry
	isReExecutingHistory bool // Flag to prevent duplicate history entries
	
	// Layout
	width  int
	height int
	
	// Error
	err error
}

type ShellState int

const (
	StateInput ShellState = iota
	StateLoading
	StateShowingContext
	StateConfirmingCommand
	StateExecuting
	StateShowingResult
	StateViewingHistory
	StateError
)

type HistoryEntry struct {
	Prompt      string
	Response    string
	Mode        domain.QueryMode
	Success     bool
	Command     string // The generated command (for re-execution)
	Result      string // The execution result
	Timestamp   string // When it was executed
}

type ExecutionResult struct {
	Output  string
	Success bool
	Error   string
}

// NewSimpleShell creates a new simplified shell
func NewSimpleShell(connectionName string, shellService services.ShellServiceInterface) *SimpleShell {
	input := textinput.New()
	input.Placeholder = "What would you like to do? (Tab: switch mode, F1: context, F2: history)"
	input.Focus()
	input.CharLimit = 500

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	shell := &SimpleShell{
		connectionName: connectionName,
		shellService:   shellService,
		input:          input,
		spinner:        s,
		state:          StateInput,
		mode:           domain.QueryModeCommand,
		history:        []HistoryEntry{},
	}
	
	// Start with empty history - will be populated as user interacts
	
	return shell
}

func (m *SimpleShell) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
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
			// For answers, show the result first, then add to history when user presses Enter
			m.result = &ExecutionResult{
				Output:  m.response.Answer,
				Success: true,
				Error:   "",
			}
			m.state = StateShowingResult
		}
		return m, nil

	case executionResultMsg:
		m.result = &ExecutionResult{
			Output:  msg.output,
			Success: msg.success,
			Error:   msg.error,
		}
		m.state = StateShowingResult
		// Only add to history if this is not a re-execution from history
		if !m.isReExecutingHistory {
			m.addToHistory(m.loadingPrompt, msg.output, msg.success)
		}
		// Reset the flag
		m.isReExecutingHistory = false
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.state = StateError
		return m, nil
	}

	// Update spinner for loading states
	if m.state == StateLoading || m.state == StateExecuting {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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
	case StateLoading, StateExecuting:
		return m.handleLoadingKeys(msg)
	case StateShowingContext, StateError:
		return m.handleViewKeys(msg)
	case StateConfirmingCommand:
		return m.handleConfirmKeys(msg)
	case StateShowingResult:
		return m.handleResultKeys(msg)
	case StateViewingHistory:
		return m.handleHistoryKeys(msg)
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

	case tea.KeyF2:
		if len(m.history) > 0 {
			m.state = StateViewingHistory
			m.historyIndex = len(m.history) - 1 // Start with most recent
			return m, nil
		}
		return m, nil

	case tea.KeyTab:
		m.toggleMode()
		return m, nil

	case tea.KeyEnter:
		prompt := strings.TrimSpace(m.input.Value())
		if prompt != "" {
			m.loadingPrompt = prompt
			m.state = StateLoading
			return m, tea.Batch(m.spinner.Tick, m.processInput(prompt))
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

func (m *SimpleShell) handleLoadingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	return m, nil
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
			m.state = StateExecuting
			return m, tea.Batch(m.spinner.Tick, m.executeCommand(m.response.Command))
		}
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeyEsc:
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeyLeft:
		if m.confirmScrollOffset > 0 {
			m.confirmScrollOffset = max(m.confirmScrollOffset - 5, 0)
		}
		return m, nil
	case tea.KeyRight:
		if m.confirmScrollOffset < m.confirmMaxScroll {
			m.confirmScrollOffset = min(m.confirmScrollOffset + 5, m.confirmMaxScroll)
		}
		return m, nil
	case tea.KeyUp:
		if m.confirmVertOffset > 0 {
			m.confirmVertOffset--
		}
		return m, nil
	case tea.KeyDown:
		if m.confirmVertOffset < m.confirmMaxVertical {
			m.confirmVertOffset++
		}
		return m, nil
	case tea.KeyPgUp:
		m.confirmVertOffset = max(0, m.confirmVertOffset-3)
		return m, nil
	case tea.KeyPgDown:
		m.confirmVertOffset = min(m.confirmMaxVertical, m.confirmVertOffset+3)
		return m, nil
	case tea.KeyHome:
		m.confirmScrollOffset = 0
		m.confirmVertOffset = 0
		return m, nil
	case tea.KeyEnd:
		m.confirmScrollOffset = m.confirmMaxScroll
		m.confirmVertOffset = m.confirmMaxVertical
		return m, nil
	}
	return m, nil
}

func (m *SimpleShell) handleResultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter, tea.KeyEsc:
		// Add to history if this was an answer mode response
		if m.response != nil && m.response.Type == services.ResponseTypeAnswer {
			m.addToHistory(m.loadingPrompt, m.response.Answer, true)
		}
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeyLeft:
		if m.scrollOffset > 0 {
			m.scrollOffset = max(m.scrollOffset - 5, 0) // Scroll left by 5 characters
		}
		return m, nil
	case tea.KeyRight:
		if m.scrollOffset < m.maxScroll {
			m.scrollOffset = min(m.scrollOffset + 5, m.maxScroll) // Scroll right by 5 characters
		}
		return m, nil
	case tea.KeyUp:
		if m.verticalOffset > 0 {
			m.verticalOffset--
		}
		return m, nil
	case tea.KeyDown:
		if m.verticalOffset < m.maxVertical {
			m.verticalOffset++
		}
		return m, nil
	case tea.KeyPgUp:
		m.verticalOffset = max(0, m.verticalOffset-5)
		return m, nil
	case tea.KeyPgDown:
		m.verticalOffset = min(m.maxVertical, m.verticalOffset+5)
		return m, nil
	case tea.KeyHome:
		m.scrollOffset = 0
		m.verticalOffset = 0
		return m, nil
	case tea.KeyEnd:
		m.scrollOffset = m.maxScroll
		m.verticalOffset = m.maxVertical
		return m, nil
	}
	return m, nil
}

func (m *SimpleShell) handleHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeyUp:
		if m.historyIndex > 0 {
			m.historyIndex--
		}
		return m, nil
	case tea.KeyDown:
		if m.historyIndex < len(m.history)-1 {
			m.historyIndex++
		}
		return m, nil
	case tea.KeyEnter:
		// Re-execute the selected history item
		if m.historyIndex >= 0 && m.historyIndex < len(m.history) {
			selected := m.history[m.historyIndex]
			// Move this item to the top of the history (most recent)
			m.moveHistoryItemToTop(m.historyIndex)
			
			if selected.Command != "" {
				// Execute the stored command directly
				// Set loadingPrompt from history so it doesn't create empty entry
				m.loadingPrompt = selected.Prompt
				m.isReExecutingHistory = true
				m.state = StateExecuting
				return m, tea.Batch(m.spinner.Tick, m.executeCommand(selected.Command))
			} else if selected.Mode == domain.QueryModeAnswer {
				// Show the stored answer
				m.result = &ExecutionResult{
					Output:  selected.Result,
					Success: selected.Success,
					Error:   "",
				}
				// Clear response to prevent adding to history again when returning
				m.response = nil
				m.loadingPrompt = ""
				m.state = StateShowingResult
				return m, nil
			}
		}
		m.resetToInput()
		return m, textinput.Blink
	case tea.KeySpace:
		// View result without re-executing
		if m.historyIndex >= 0 && m.historyIndex < len(m.history) {
			selected := m.history[m.historyIndex]
			// Move this item to the top of the history (most recent)
			m.moveHistoryItemToTop(m.historyIndex)
			
			m.result = &ExecutionResult{
				Output:  selected.Result,
				Success: selected.Success,
				Error:   "",
			}
			// Clear response to prevent adding to history again when returning
			m.response = nil
			m.loadingPrompt = ""
			m.state = StateShowingResult
			return m, nil
		}
		return m, nil
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
	case StateLoading:
		content.WriteString(m.renderLoading())
	case StateShowingContext:
		content.WriteString(m.renderContext())
	case StateConfirmingCommand:
		content.WriteString(m.renderConfirmation())
	case StateExecuting:
		content.WriteString(m.renderExecuting())
	case StateShowingResult:
		content.WriteString(m.renderResult())
	case StateViewingHistory:
		content.WriteString(m.renderHistoryBrowser())
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
	content.WriteString(helpStyle.Render("Tab: switch mode • F1: context • F2: history • ↑↓: navigate • Enter: submit"))

	return content.String()
}

func (m *SimpleShell) renderLoading() string {
	var content strings.Builder
	
	// Show different messages based on mode
	var loadingText string
	if m.mode == domain.QueryModeCommand {
		loadingText = fmt.Sprintf("Generating command for: %s", m.loadingPrompt)
	} else {
		loadingText = fmt.Sprintf("Generating answer for: %s", m.loadingPrompt)
	}
	
	content.WriteString(fmt.Sprintf("%s %s\n\n", m.spinner.View(), loadingText))
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString(helpStyle.Render("Please wait... Press Ctrl+C to cancel"))
	
	return content.String()
}

func (m *SimpleShell) renderExecuting() string {
	var content strings.Builder
	
	if m.response != nil {
		commandStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
		
		content.WriteString("Executing command:\n\n")
		content.WriteString(commandStyle.Render(m.response.Command))
		content.WriteString("\n\n")
	}
	
	content.WriteString(fmt.Sprintf("%s Executing...\n\n", m.spinner.View()))
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString(helpStyle.Render("Please wait... Press Ctrl+C to cancel"))
	
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

	// Show first few items
	for i, item := range m.context.Items {
		if i >= 10 {
			content.WriteString(fmt.Sprintf("... and %d more items\n", len(m.context.Items)-10))
			break
		}
		content.WriteString(fmt.Sprintf("%s %s\n", item.Icon, item.Name))
	}

	content.WriteString("\nPress Enter to return")
	return content.String()
}

func (m *SimpleShell) renderConfirmation() string {
	if m.response == nil {
		return "No command to confirm"
	}

	var content strings.Builder
	content.WriteString("Generated command:\n\n")
	
	// Handle scrollable command output
	scrolledCommand := m.renderScrollableCommand(m.response.Command)
	
	commandStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	content.WriteString(commandStyle.Render(scrolledCommand))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Safety Level: %s", m.response.Safety))
	
	// Show scroll instructions if there's content to scroll
	if m.confirmMaxScroll > 0 || m.confirmMaxVertical > 0 {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		
		var scrollInfo strings.Builder
		scrollInfo.WriteString("\n")
		
		// Horizontal scroll indicator
		if m.confirmMaxScroll > 0 {
			scrollPercent := int(float64(m.confirmScrollOffset) / float64(m.confirmMaxScroll) * 100)
			if scrollPercent > 100 {
				scrollPercent = 100
			}
			
			leftArrow := "◀"
			rightArrow := "▶"
			if m.confirmScrollOffset == 0 {
				leftArrow = "◁"
			}
			if m.confirmScrollOffset >= m.confirmMaxScroll {
				rightArrow = "▷"
			}
			
			scrollInfo.WriteString(fmt.Sprintf("%s %d%% %s ", leftArrow, scrollPercent, rightArrow))
		}
		
		// Vertical scroll indicator
		if m.confirmMaxVertical > 0 {
			verticalPercent := int(float64(m.confirmVertOffset) / float64(m.confirmMaxVertical) * 100)
			if verticalPercent > 100 {
				verticalPercent = 100
			}
			
			upArrow := "▲"
			downArrow := "▼"
			if m.confirmVertOffset == 0 {
				upArrow = "△"
			}
			if m.confirmVertOffset >= m.confirmMaxVertical {
				downArrow = "▽"
			}
			
			if m.confirmMaxScroll > 0 {
				scrollInfo.WriteString("• ")
			}
			scrollInfo.WriteString(fmt.Sprintf("%s %d%% %s", upArrow, verticalPercent, downArrow))
		}
		
		// Add navigation instructions
		var instructions strings.Builder
		if m.confirmMaxScroll > 0 && m.confirmMaxVertical > 0 {
			instructions.WriteString("←→: horizontal • ↑↓: vertical • PgUp/PgDn: page • Home/End: jump")
		} else if m.confirmMaxScroll > 0 {
			instructions.WriteString("←→: scroll • Home/End: jump")
		} else if m.confirmMaxVertical > 0 {
			instructions.WriteString("↑↓: scroll • PgUp/PgDn: page • Home/End: jump")
		}
		instructions.WriteString(" • Enter: execute • Esc: cancel")
		
		scrollInfo.WriteString(" • ")
		scrollInfo.WriteString(helpStyle.Render(instructions.String()))
		
		content.WriteString(scrollInfo.String())
	} else {
		content.WriteString("\n\nPress Enter to execute • Esc to cancel")
	}
	
	return content.String()
}

func (m *SimpleShell) renderResult() string {
	if m.result == nil {
		return "No result available\n\nPress Enter to continue"
	}
	
	var content strings.Builder
	
	if m.result.Success {
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
		content.WriteString(successStyle.Render("✅ Execution completed successfully"))
	} else {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		content.WriteString(errorStyle.Render("❌ Execution failed"))
		if m.result.Error != "" {
			content.WriteString(fmt.Sprintf("\n\nError: %s", m.result.Error))
		}
	}
	
	content.WriteString("\n\n")
	
	if m.result.Output != "" {
		content.WriteString("Output:\n\n")
		
		// Handle scrollable table output
		scrolledOutput := m.renderScrollableOutput(m.result.Output)
		
		resultStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
		
		content.WriteString(resultStyle.Render(scrolledOutput))
	}
	
	// Show scroll instructions if there's content to scroll
	if m.maxScroll > 0 || m.maxVertical > 0 {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		
		var scrollInfo strings.Builder
		scrollInfo.WriteString("\n")
		
		// Horizontal scroll indicator
		if m.maxScroll > 0 {
			scrollPercent := int(float64(m.scrollOffset) / float64(m.maxScroll) * 100)
			if scrollPercent > 100 {
				scrollPercent = 100
			}
			
			leftArrow := "◀"
			rightArrow := "▶"
			if m.scrollOffset == 0 {
				leftArrow = "◁" // Inactive left arrow
			}
			if m.scrollOffset >= m.maxScroll {
				rightArrow = "▷" // Inactive right arrow
			}
			
			scrollInfo.WriteString(fmt.Sprintf("%s %d%% %s ", leftArrow, scrollPercent, rightArrow))
		}
		
		// Vertical scroll indicator
		if m.maxVertical > 0 {
			verticalPercent := int(float64(m.verticalOffset) / float64(m.maxVertical) * 100)
			if verticalPercent > 100 {
				verticalPercent = 100
			}
			
			upArrow := "▲"
			downArrow := "▼"
			if m.verticalOffset == 0 {
				upArrow = "△" // Inactive up arrow
			}
			if m.verticalOffset >= m.maxVertical {
				downArrow = "▽" // Inactive down arrow
			}
			
			if m.maxScroll > 0 {
				scrollInfo.WriteString("• ")
			}
			scrollInfo.WriteString(fmt.Sprintf("%s %d%% %s", upArrow, verticalPercent, downArrow))
		}
		
		// Add navigation instructions
		var instructions strings.Builder
		if m.maxScroll > 0 && m.maxVertical > 0 {
			instructions.WriteString("←→: horizontal • ↑↓: vertical (sticky headers) • PgUp/PgDn: page • Home/End: jump")
		} else if m.maxScroll > 0 {
			instructions.WriteString("←→: scroll • Home/End: jump")
		} else if m.maxVertical > 0 {
			instructions.WriteString("↑↓: scroll (sticky headers) • PgUp/PgDn: page • Home/End: jump")
		}
		instructions.WriteString(" • Enter: continue")
		
		scrollInfo.WriteString(" • ")
		scrollInfo.WriteString(helpStyle.Render(instructions.String()))
		
		content.WriteString(scrollInfo.String())
	} else {
		content.WriteString("\n\nPress Enter to continue")
	}
	
	return content.String()
}

func (m *SimpleShell) renderScrollableOutput(output string) string {
	lines := strings.Split(output, "\n")
	availableWidth := m.width - 8 // Account for padding and border
	availableHeight := m.height - 12 // Account for header, instructions, etc.
	
	if availableWidth < 20 {
		availableWidth = 80 // Fallback
	}
	if availableHeight < 5 {
		availableHeight = 10 // Fallback
	}
	
	maxLineWidth := 0
	
	// Calculate max line width to determine horizontal scroll bounds
	for _, line := range lines {
		if len(line) > maxLineWidth {
			maxLineWidth = len(line)
		}
	}
	
	// Update horizontal scroll bounds
	if maxLineWidth > availableWidth {
		m.maxScroll = maxLineWidth - availableWidth
	} else {
		m.maxScroll = 0
		m.scrollOffset = 0
	}
	
	// Detect table headers (sticky headers)
	headerLines, dataLines := m.separateHeaderAndData(lines)
	headerHeight := len(headerLines)
	
	// Reserve space for headers in available height
	dataAvailableHeight := availableHeight - headerHeight
	if dataAvailableHeight < 1 {
		dataAvailableHeight = 1 // Always show at least one data row
	}
	
	// Update vertical scroll bounds based on data lines only
	if len(dataLines) > dataAvailableHeight {
		m.maxVertical = len(dataLines) - dataAvailableHeight
	} else {
		m.maxVertical = 0
		m.verticalOffset = 0
	}
	
	// Select visible data lines (headers are always shown)
	var visibleDataLines []string
	if len(dataLines) > 0 {
		startLine := m.verticalOffset
		endLine := m.verticalOffset + dataAvailableHeight
		if endLine > len(dataLines) {
			endLine = len(dataLines)
		}
		
		visibleDataLines = dataLines[startLine:endLine]
	}
	
	// Combine headers + visible data
	var visibleLines []string
	visibleLines = append(visibleLines, headerLines...)
	visibleLines = append(visibleLines, visibleDataLines...)
	
	// Apply horizontal scrolling to all visible lines
	var scrolledLines []string
	for _, line := range visibleLines {
		if len(line) <= availableWidth {
			// Line fits entirely, no horizontal scrolling needed
			scrolledLines = append(scrolledLines, line)
		} else {
			// Apply horizontal scrolling
			start := m.scrollOffset
			end := m.scrollOffset + availableWidth
			
			if start >= len(line) {
				scrolledLines = append(scrolledLines, "")
			} else if end >= len(line) {
				scrolledLines = append(scrolledLines, line[start:])
			} else {
				scrolledLines = append(scrolledLines, line[start:end])
			}
		}
	}
	
	return strings.Join(scrolledLines, "\n")
}

func (m *SimpleShell) renderScrollableCommand(command string) string {
	lines := strings.Split(command, "\n")
	availableWidth := m.width - 10 // Account for padding and border
	availableHeight := m.height - 15 // Account for header, instructions, etc.
	
	if availableWidth < 20 {
		availableWidth = 60 // Fallback for command display
	}
	if availableHeight < 3 {
		availableHeight = 5 // Fallback
	}
	
	maxLineWidth := 0
	
	// Calculate max line width to determine horizontal scroll bounds
	for _, line := range lines {
		if len(line) > maxLineWidth {
			maxLineWidth = len(line)
		}
	}
	
	// Update horizontal scroll bounds for confirmation
	if maxLineWidth > availableWidth {
		m.confirmMaxScroll = maxLineWidth - availableWidth
	} else {
		m.confirmMaxScroll = 0
		m.confirmScrollOffset = 0
	}
	
	// Update vertical scroll bounds for confirmation
	if len(lines) > availableHeight {
		m.confirmMaxVertical = len(lines) - availableHeight
	} else {
		m.confirmMaxVertical = 0
		m.confirmVertOffset = 0
	}
	
	// Apply vertical scrolling first (select which lines to show)
	startLine := m.confirmVertOffset
	endLine := m.confirmVertOffset + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	visibleLines := lines[startLine:endLine]
	
	// Apply horizontal scrolling to visible lines
	var scrolledLines []string
	for _, line := range visibleLines {
		if len(line) <= availableWidth {
			// Line fits entirely, no horizontal scrolling needed
			scrolledLines = append(scrolledLines, line)
		} else {
			// Apply horizontal scrolling
			start := m.confirmScrollOffset
			end := m.confirmScrollOffset + availableWidth
			
			if start >= len(line) {
				scrolledLines = append(scrolledLines, "")
			} else if end >= len(line) {
				scrolledLines = append(scrolledLines, line[start:])
			} else {
				scrolledLines = append(scrolledLines, line[start:end])
			}
		}
	}
	
	return strings.Join(scrolledLines, "\n")
}

func (m *SimpleShell) separateHeaderAndData(lines []string) ([]string, []string) {
	if len(lines) == 0 {
		return []string{}, []string{}
	}
	
	// Look for standard ASCII table pattern:
	// +----+------+------+
	// | ID | NAME | EMAIL|  <- header row
	// +----+------+------+
	// | 1  | John | john@|  <- data starts here
	// | 2  | Jane | jane@|
	
	headerEndIndex := 0
	foundHeaderRow := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// Check if this is a border line (contains + and -)
		isBorderLine := strings.Contains(trimmed, "+") && strings.Contains(trimmed, "-")
		
		// Check if this looks like a header row (contains | but not border, and has typical header words)
		isLikelyHeaderRow := strings.Contains(trimmed, "|") && 
			!isBorderLine &&
			i < 5 && // Headers should be in first few lines
			(strings.Contains(strings.ToUpper(trimmed), "ID") ||
			 strings.Contains(strings.ToUpper(trimmed), "NAME") ||
			 strings.Contains(strings.ToUpper(trimmed), "EMAIL") ||
			 strings.Contains(strings.ToUpper(trimmed), "FIRST_NAME") ||
			 strings.Contains(strings.ToUpper(trimmed), "LAST_NAME") ||
			 strings.Contains(strings.ToUpper(trimmed), "DEPARTMENT") ||
			 strings.Contains(strings.ToUpper(trimmed), "PHONE") ||
			 strings.Contains(strings.ToUpper(trimmed), "HIRE_DATE") ||
			 strings.Contains(strings.ToUpper(trimmed), "SALARY"))
		
		if isLikelyHeaderRow {
			foundHeaderRow = true
		}
		
		// If we found a header row, continue including lines until we hit a border followed by data
		if foundHeaderRow {
			if isBorderLine {
				// This border line after header row marks end of header section
				headerEndIndex = i + 1
				break
			}
		}
	}
	
	// Fallback: if no clear pattern found, assume first 3 lines are headers
	if headerEndIndex == 0 && len(lines) > 3 {
		// Look for at least one border line in first 3 lines to confirm it's a table
		hasTableStructure := false
		for i := 0; i < min(3, len(lines)); i++ {
			if strings.Contains(lines[i], "+") && strings.Contains(lines[i], "-") {
				hasTableStructure = true
				break
			}
		}
		
		if hasTableStructure {
			headerEndIndex = 3
		} else {
			// No table structure detected, treat everything as data
			headerEndIndex = 0
		}
	}
	
	// Split lines into header and data sections
	var headerLines []string
	var dataLines []string
	
	if headerEndIndex > 0 && headerEndIndex < len(lines) {
		headerLines = lines[:headerEndIndex]
		dataLines = lines[headerEndIndex:]
	} else {
		// No headers detected, treat all as data
		headerLines = []string{}
		dataLines = lines
	}
	
	return headerLines, dataLines
}

func (m *SimpleShell) createDemoWideTable() string {
	// Create a demo table with many columns and rows to test both horizontal and vertical scrolling
	return `+----+------------------+------------------+------------------+------------------+------------------+------------------+------------------+
| ID | FIRST_NAME       | LAST_NAME        | EMAIL_ADDRESS    | PHONE_NUMBER     | DEPARTMENT       | HIRE_DATE        | ANNUAL_SALARY    |
+----+------------------+------------------+------------------+------------------+------------------+------------------+------------------+
| 1  | John             | Smith            | john@company.com | +1-555-123-4567  | Engineering      | 2020-01-15       | $95,000          |
| 2  | Sarah            | Johnson          | sarah@company.com| +1-555-234-5678  | Marketing        | 2019-03-22       | $78,000          |
| 3  | Michael          | Brown            | mike@company.com | +1-555-345-6789  | Sales            | 2021-07-10       | $65,000          |
| 4  | Emily            | Davis            | emily@company.com| +1-555-456-7890  | Engineering      | 2020-11-05       | $88,000          |
| 5  | Robert           | Wilson           | rob@company.com  | +1-555-567-8901  | HR               | 2018-12-01       | $72,000          |
| 6  | Lisa             | Anderson         | lisa@company.com | +1-555-678-9012  | Engineering      | 2021-08-15       | $92,000          |
| 7  | David            | Thompson         | david@company.com| +1-555-789-0123  | Marketing        | 2020-05-20       | $76,000          |
| 8  | Jennifer         | Garcia           | jen@company.com  | +1-555-890-1234  | Sales            | 2019-09-10       | $68,000          |
| 9  | William          | Martinez         | will@company.com | +1-555-901-2345  | Engineering      | 2021-01-25       | $97,000          |
| 10 | Amanda           | Robinson         | amanda@company.com| +1-555-012-3456  | HR               | 2018-06-14       | $74,000          |
| 11 | Christopher      | Clark            | chris@company.com| +1-555-123-4567  | Sales            | 2020-10-03       | $71,000          |
| 12 | Michelle         | Rodriguez        | michelle@company.com| +1-555-234-5678| Marketing        | 2019-11-28       | $79,000          |
| 13 | James            | Lewis            | james@company.com| +1-555-345-6789  | Engineering      | 2021-04-12       | $94,000          |
| 14 | Patricia         | Lee              | pat@company.com  | +1-555-456-7890  | HR               | 2018-08-07       | $73,000          |
| 15 | Daniel           | Walker           | dan@company.com  | +1-555-567-8901  | Sales            | 2020-12-18       | $69,000          |
+----+------------------+------------------+------------------+------------------+------------------+------------------+------------------+`
}

func (m *SimpleShell) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	return fmt.Sprintf("%s\n\nPress Enter to continue", 
		errorStyle.Render(fmt.Sprintf("Error: %s", m.err)))
}

func (m *SimpleShell) renderHistoryBrowser() string {
	if len(m.history) == 0 {
		return "No history available\n\nPress Esc to return"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render("📚 Command History"))
	content.WriteString("\n\n")

	// Show history items with selection
	for i, item := range m.history {
		var prefix string
		if i == m.historyIndex {
			prefix = "▶ "
		} else {
			prefix = "  "
		}

		// Format the entry
		icon := "✓"
		if !item.Success {
			icon = "✗"
		}
		
		modeIcon := "🔧"
		if item.Mode == domain.QueryModeAnswer {
			modeIcon = "❓"
		}

		entryStyle := lipgloss.NewStyle()
		if i == m.historyIndex {
			entryStyle = entryStyle.Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("255")).
				Bold(true)
		} else {
			entryStyle = entryStyle.Foreground(lipgloss.Color("252"))
		}

		timestamp := item.Timestamp
		if timestamp == "" {
			timestamp = "unknown"
		}

		line := fmt.Sprintf("%s%s %s %s", prefix, icon, modeIcon, truncate(item.Prompt, 60))
		content.WriteString(entryStyle.Render(line))
		content.WriteString("\n")

		// Show timestamp and result preview for selected item
		if i == m.historyIndex {
			detailStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 0, 0, 4)
			
			details := fmt.Sprintf("Time: %s", timestamp)
			if item.Command != "" {
				details += fmt.Sprintf("\nCommand: %s", truncate(item.Command, 50))
			}
			if item.Result != "" {
				details += fmt.Sprintf("\nResult: %s", truncate(item.Result, 50))
			}
			
			content.WriteString(detailStyle.Render(details))
			content.WriteString("\n")
		}
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: navigate • Enter: re-execute • Space: view result • Esc: back"))

	return content.String()
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
	m.result = nil
	m.loadingPrompt = ""
	m.scrollOffset = 0
	m.maxScroll = 0
	m.verticalOffset = 0
	m.maxVertical = 0
	m.confirmScrollOffset = 0
	m.confirmMaxScroll = 0
	m.confirmVertOffset = 0
	m.confirmMaxVertical = 0
	m.historyIndex = 0
	m.selectedHistory = nil
	m.isReExecutingHistory = false
}

func (m *SimpleShell) addToHistory(prompt, response string, success bool) {
	// Extract command and result based on context
	command := ""
	result := response
	
	if m.response != nil && m.response.Command != "" {
		command = m.response.Command
	}
	
	now := time.Now()
	timestamp := now.Format("15:04")
	
	m.history = append(m.history, HistoryEntry{
		Prompt:    prompt,
		Response:  response,
		Mode:      m.mode,
		Success:   success,
		Command:   command,
		Result:    result,
		Timestamp: timestamp,
	})

	// Keep only last 10 items
	if len(m.history) > 10 {
		m.history = m.history[1:]
	}
}

// moveHistoryItemToTop moves the history item at the given index to the most recent position
func (m *SimpleShell) moveHistoryItemToTop(index int) {
	if index < 0 || index >= len(m.history) {
		return
	}
	
	// Get the item to move
	item := m.history[index]
	
	// Update timestamp to reflect when it was accessed
	now := time.Now()
	item.Timestamp = now.Format("15:04")
	
	// Remove it from current position
	m.history = append(m.history[:index], m.history[index+1:]...)
	
	// Add it to the end (most recent position)
	m.history = append(m.history, item)
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
		// Get the original connection to execute the command
		connection, err := config.GetConnectionByName(m.connectionName)
		if err != nil {
			return executionResultMsg{
				output:  "",
				success: false,
				error:   fmt.Sprintf("Connection not found: %v", err),
			}
		}

		// Get the connection implementation
		connImpl, err := conn.GetConnection(connection)
		if err != nil {
			return executionResultMsg{
				output:  "",
				success: false,
				error:   fmt.Sprintf("Failed to get connection: %v", err),
			}
		}

		// Execute the command
		result, err := connImpl.ExecuteCommand(command)
		if err != nil {
			return executionResultMsg{
				output:  "",
				success: false,
				error:   fmt.Sprintf("Execution failed: %v", err),
			}
		}

		// Format the result
		var output string
		if len(result) > 0 {
			// Try to format as table if it's a database connection
			if connection.Type.GetMainType() == "Database" {
				if dbConn, ok := connImpl.(interface{ FormatResultAsTable([]byte) (string, error) }); ok {
					if formatted, err := dbConn.FormatResultAsTable(result); err == nil {
						output = formatted
					} else {
						output = string(result)
					}
				} else {
					output = string(result)
				}
			} else {
				output = string(result)
			}
		} else {
			// Create a demo wide table for testing scrolling
			if connection.Type.GetMainType() == "Database" {
				output = m.createDemoWideTable()
			} else {
				output = "Command executed successfully (no output)"
			}
		}

		return executionResultMsg{
			output:  output,
			success: true,
			error:   "",
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

type executionResultMsg struct {
	output  string
	success bool
	error   string
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