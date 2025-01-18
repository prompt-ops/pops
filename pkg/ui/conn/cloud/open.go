package cloud

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	config "github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
	"github.com/prompt-ops/pops/pkg/ui"
)

// Styles
var (
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

const (
	stepSelectConnection step = iota
	stepOpenSpinner
	stepOpenDone
)

// Message types
type (
	// Sent when our spinner is done
	doneSpinnerMsg struct{}
)

// model defines the state of the UI
type model struct {
	currentStep step
	cursor      int
	connections []conn.Connection
	selected    conn.Connection
	err         error
	spinner     spinner.Model
}

// NewOpenModel initializes the open model for Cloud connections
func NewOpenModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return model{
		currentStep: stepSelectConnection,
		spinner:     sp,
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadConnectionsCmd(),
	)
}

// loadConnectionsCmd fetches existing cloud connections
func (m model) loadConnectionsCmd() tea.Cmd {
	return func() tea.Msg {
		cloudConnections, err := config.GetConnectionsByType(conn.ConnectionTypeCloud)
		if err != nil {
			return err
		}
		if len(cloudConnections) == 0 {
			return fmt.Errorf("no cloud connections found")
		}
		return connectionsMsg{
			connections: cloudConnections,
		}
	}
}

type connectionsMsg struct {
	connections []conn.Connection
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentStep {
	case stepSelectConnection:
		switch msg := msg.(type) {
		case connectionsMsg:
			m.connections = msg.connections
			return m, nil
		case error:
			m.err = msg
			m.currentStep = stepOpenDone
			return m, nil
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.connections)-1 {
					m.cursor++
				}
			case "enter":
				m.selected = m.connections[m.cursor]
				m.currentStep = stepOpenSpinner
				return m, tea.Batch(
					m.spinner.Tick,
					transitionCmd(m.selected),
				)
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case stepOpenSpinner:
		switch msg := msg.(type) {
		case ui.TransitionToShellMsg:
			return m, tea.Quit
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case doneSpinnerMsg:
			m.currentStep = stepOpenDone
			return m, nil
		}

	case stepOpenDone:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}

	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func transitionCmd(conn conn.Connection) tea.Cmd {
	return func() tea.Msg {
		return ui.TransitionToShellMsg{
			Connection: conn,
		}
	}
}

func (m model) View() string {
	// Clear the terminal before rendering the UI
	clearScreen := "\033[H\033[2J"

	switch m.currentStep {
	case stepSelectConnection:
		s := titleStyle.Render("Select a Cloud Connection (↑/↓, Enter to open):")
		s += "\n\n"
		for i, conn := range m.connections {
			cursor := "  "
			if i == m.cursor {
				cursor = "→ "
				s += selectedStyle.Render(fmt.Sprintf("%s%s", cursor, conn.Name)) + "\n"
				continue
			}
			s += unselectedStyle.Render(fmt.Sprintf("%s%s", cursor, conn.Name)) + "\n"
		}
		s += "\n" + helpStyle.Render(ui.QuitMessage)
		return clearScreen + s

	case stepOpenSpinner:
		return clearScreen + lipgloss.JoinHorizontal(lipgloss.Left,
			fmt.Sprintf("Opening connection '%s'...", m.selected.Name),
			m.spinner.View(),
		)

	case stepOpenDone:
		if m.err != nil {
			return clearScreen + errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress 'q' or 'esc' to quit.", m.err))
		}
		return clearScreen + lipgloss.JoinHorizontal(lipgloss.Left,
			"✅ Connection opened!",
			"\n\nPress 'Enter' or 'q'/'esc' to exit.",
		)
	default:
		return clearScreen
	}
}
