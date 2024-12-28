package kubernetes

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/common"
	config "github.com/prompt-ops/pops/config"
	"github.com/prompt-ops/pops/ui"
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

// openModel defines the state of the UI
type openModel struct {
	currentStep step
	cursor      int
	connections []common.Connection
	selected    common.Connection
	err         error

	// Spinner for transitions
	spinner spinner.Model
}

// NewOpenModel initializes the open openModel for Kubernetes connections
func NewOpenModel() *openModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &openModel{
		currentStep: stepSelectConnection,
		spinner:     sp,
	}
}

// Init initializes the openModel
func (m *openModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadConnectionsCmd(),
	)
}

// loadConnectionsCmd fetches existing Kubernetes connections
func (m *openModel) loadConnectionsCmd() tea.Cmd {
	return func() tea.Msg {
		k8sConnections, err := config.GetConnectionsByType(common.ConnectionTypeKubernetes)
		if err != nil {
			return err
		}
		if len(k8sConnections) == 0 {
			return fmt.Errorf("no Kubernetes connections found")
		}
		return connectionsMsg{
			connections: k8sConnections,
		}
	}
}

// connectionsMsg holds the list of Kubernetes connections
type connectionsMsg struct {
	connections []common.Connection
}

// Update handles incoming messages and updates the openModel accordingly
func (m *openModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			switch msg.String() {
			case "enter", "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	// Always update the spinner
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func transitionCmd(conn common.Connection) tea.Cmd {
	return func() tea.Msg {
		return ui.TransitionToShellMsg{
			Connection: conn,
		}
	}
}

func (m *openModel) View() string {
	switch m.currentStep {
	case stepSelectConnection:
		s := titleStyle.Render("Select a Kubernetes Connection (↑/↓, Enter to open):")
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
		s += "\n" + helpStyle.Render("Press 'q' or 'esc' or Ctrl+C to quit.")
		return s

	case stepOpenSpinner:
		return fmt.Sprintf("Opening connection '%s'... %s", m.selected.Name, m.spinner.View())

	case stepOpenDone:
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress 'q' or 'esc' to quit.", m.err))
		}
		return fmt.Sprintf("✅ Connection opened!\n\nPress 'Enter' or 'q'/'esc' to exit.")

	default:
		return ""
	}
}
