package k8s

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	config "github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
	"github.com/prompt-ops/pops/pkg/ui"
)

// Styles
var (
	outputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

const (
	stepSelectContext step = iota
	stepEnterConnectionName
	stepCreateSpinner
	stepCreateDone
)

type (
	doneWaitingMsg struct {
		Connection conn.Connection
	}

	contextsMsg struct {
		contexts []string
	}

	errMsg struct {
		err error
	}
)

// createModel defines the state of the UI
type createModel struct {
	currentStep step
	cursor      int
	contexts    []string
	selectedCtx string
	input       textinput.Model
	err         error

	// Spinner for the 2-second wait
	spinner spinner.Model

	connection conn.Connection
}

// NewCreateModel initializes the createModel for Kubernetes
func NewCreateModel() *createModel {
	ti := textinput.New()
	ti.Placeholder = ui.EnterConnectionNameMessage
	ti.CharLimit = 256
	ti.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &createModel{
		currentStep: stepSelectContext,
		input:       ti,
		spinner:     sp,
	}
}

// Init initializes the createModel
func (m *createModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadContextsCmd(),
	)
}

// loadContextsCmd fetches available Kubernetes contexts
func (m *createModel) loadContextsCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("kubectl", "config", "get-contexts", "--output=name").Output()
		if err != nil {
			return errMsg{err}
		}
		contextList := strings.Split(strings.TrimSpace(string(out)), "\n")
		return contextsMsg{contexts: contextList}
	}
}

// waitTwoSecondsCmd simulates a delay for saving the connection asynchronously
func waitTwoSecondsCmd(conn conn.Connection) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return doneWaitingMsg{
			Connection: conn,
		}
	})
}

// Update handles incoming messages and updates the createModel accordingly
func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentStep {
	case stepSelectContext:
		switch msg := msg.(type) {
		case contextsMsg:
			m.contexts = msg.contexts
			if len(m.contexts) == 0 {
				m.err = fmt.Errorf("no Kubernetes contexts found")
				m.currentStep = stepCreateSpinner
				return m, nil
			}
			// Clear any previous errors when successfully loading contexts
			m.err = nil
			return m, nil
		case errMsg:
			m.err = msg.err
			m.currentStep = stepCreateSpinner
			return m, nil
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.contexts)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.contexts) > 0 && m.cursor >= 0 && m.cursor < len(m.contexts) {
					m.selectedCtx = m.contexts[m.cursor]
					m.currentStep = stepEnterConnectionName
					m.input.Focus()

					// Clear any previous errors when moving to a new step
					m.err = nil

					return m, nil
				}
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}

		// Update spinner if it's running in stepSelectContext
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepEnterConnectionName:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.input.Value())
				if name == "" {
					m.err = fmt.Errorf("connection name can't be empty")
					return m, nil
				}
				if config.CheckIfNameExists(name) {
					m.err = fmt.Errorf("connection name already exists")
					return m, nil
				}

				connection := conn.NewKubernetesConnection(name, m.selectedCtx)
				if err := config.SaveConnection(connection); err != nil {
					m.err = err
					return m, nil
				}
				m.currentStep = stepCreateSpinner
				m.err = nil
				return m, tea.Batch(
					m.spinner.Tick,
					waitTwoSecondsCmd(connection),
				)
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case spinner.TickMsg:
			return m, nil
		}

		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stepCreateSpinner:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd

		case doneWaitingMsg:
			m.connection = msg.Connection
			m.currentStep = stepCreateDone
			m.err = nil
			return m, nil

		case errMsg:
			m.err = msg.err
			m.currentStep = stepCreateDone
			m.connection = conn.Connection{}
			return m, nil

		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}

		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepCreateDone:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				return m, func() tea.Msg {
					return ui.TransitionToShellMsg{
						Connection: m.connection,
					}
				}
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case spinner.TickMsg:
			return m, nil
		}
	}

	return m, cmd
}

func (m *createModel) View() string {
	// Clear the terminal before rendering the UI
	clearScreen := "\033[H\033[2J"

	switch m.currentStep {
	case stepSelectContext:
		s := titleStyle.Render("Select a Kubernetes context (↑/↓, Enter to confirm):")
		s += "\n\n"
		for i, ctx := range m.contexts {
			cursor := "  "
			if i == m.cursor {
				cursor = "→ "
				s += selectedStyle.Render(cursor+ctx) + "\n"
				continue
			}
			s += unselectedStyle.Render(cursor+ctx) + "\n"
		}
		s += "\n" + helpStyle.Render(ui.QuitMessage)
		return clearScreen + s

	case stepEnterConnectionName:
		s := titleStyle.Render("Enter a name for the Kubernetes connection:")
		s += "\n\n"
		if m.err != nil {
			s += errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
			s += "\n"
		}
		s += m.input.View()
		s += "\n" + helpStyle.Render(ui.QuitMessage)
		return clearScreen + s

	case stepCreateSpinner:
		return clearScreen + outputStyle.Render("Saving conn... ") + m.spinner.View()

	case stepCreateDone:
		if m.err != nil {
			return clearScreen + errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress 'Enter' or 'q'/'esc' to quit.", m.err))
		}

		return clearScreen + outputStyle.Render("✅ Kubernetes connection created!\n\nPress 'Enter' or 'q'/'esc' to exit.")

	default:
		return clearScreen
	}
}
