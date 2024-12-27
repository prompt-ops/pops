package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	config "github.com/prompt-ops/pops/config"
	"github.com/prompt-ops/pops/ui"
)

var (
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	outputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

const (
	stepSelectDriver step = iota
	stepEnterConnectionString
	stepEnterConnectionName
	stepCreateSpinner
	stepCreateDone
)

var drivers = []string{"PostgreSQL"}

type (
	doneWaitingMsg struct {
		Connection config.Connection
	}

	errMsg struct {
		err error
	}
)

type createModel struct {
	currentStep step
	cursor      int

	driver           string
	connectionString string
	connection       config.Connection

	input           textinput.Model
	connectionInput textinput.Model

	spinner spinner.Model

	err error
}

func NewCreateModel() *createModel {
	ti := textinput.New()
	ti.Placeholder = "Enter connection name..."
	ti.CharLimit = 256
	ti.Width = 30

	ci := textinput.New()
	ci.Placeholder = "Enter connection string..."
	ci.CharLimit = 512
	ci.Width = 50

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &createModel{
		currentStep:      stepSelectDriver,
		cursor:           0,
		driver:           "",
		connectionString: "",
		input:            ti,
		connectionInput:  ci,
		spinner:          sp,
		err:              nil,
	}
}

func handleQuit(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
		return tea.Quit
	}
	return nil
}

func (m *createModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch m.currentStep {
	case stepSelectDriver:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if quitCmd := handleQuit(msg); quitCmd != nil {
				return m, quitCmd
			}
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(drivers)-1 {
					m.cursor++
				}
			case "enter":
				m.driver = drivers[m.cursor]
				m.currentStep = stepEnterConnectionString
				m.err = nil
				return m, m.connectionInput.Focus()
			}
		}

	case stepEnterConnectionString:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if quitCmd := handleQuit(msg); quitCmd != nil {
				return m, quitCmd
			}
			switch msg.String() {
			case "enter":
				connStr := strings.TrimSpace(m.connectionInput.Value())
				if connStr == "" {
					m.err = fmt.Errorf("connection string can't be empty")
					return m, nil
				}
				m.connectionString = connStr
				m.currentStep = stepEnterConnectionName
				m.err = nil
				return m, m.input.Focus()
			}
		}
		m.connectionInput, cmd = m.connectionInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case stepEnterConnectionName:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if quitCmd := handleQuit(msg); quitCmd != nil {
				return m, quitCmd
			}
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

				m.connection = config.NewDatabaseConnection(name, m.driver, m.connectionString)

				if err := config.SaveConnection(m.connection); err != nil {
					m.err = err
					m.currentStep = stepCreateDone
					return m, nil
				}

				m.currentStep = stepCreateSpinner
				m.err = nil
				return m, tea.Batch(
					m.spinner.Tick,
					waitTwoSecondsCmd(m.connection),
				)
			}
		}
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

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
			m.connection = config.Connection{}
			return m, nil
		case tea.KeyMsg:
			if quitCmd := handleQuit(msg); quitCmd != nil {
				return m, quitCmd
			}
		}
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case stepCreateDone:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if quitCmd := handleQuit(msg); quitCmd != nil {
				return m, quitCmd
			}
			switch msg.String() {
			case "enter":
				return m, func() tea.Msg {
					return ui.TransitionToShellMsg{
						Connection: m.connection,
					}
				}
			}
		}
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func waitTwoSecondsCmd(conn config.Connection) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return doneWaitingMsg{
			Connection: conn,
		}
	})
}

func (m *createModel) View() string {
	switch m.currentStep {
	case stepSelectDriver:
		s := promptStyle.Render("Select a database driver (↑/↓, Enter to confirm):")
		s += "\n\n"
		for i, p := range drivers {
			cursor := "  "
			if i == m.cursor {
				cursor = "→ "
			}
			if i == m.cursor {
				s += fmt.Sprintf("%s%s\n", cursor, promptStyle.Copy().Bold(true).Render(p))
			} else {
				s += fmt.Sprintf("%s%s\n", cursor, promptStyle.Render(p))
			}
		}
		s += "\nPress 'q', 'esc', or Ctrl+C to quit."
		return s

	case stepEnterConnectionString:
		s := promptStyle.Render("Enter the connection string:")
		s += "\n\n"
		if m.err != nil {
			s += fmt.Sprintf("Error: %v", m.err)
			s += "\n\n"
		}
		s += m.connectionInput.View()
		s += "\n\nPress 'Enter' to proceed or 'q', 'esc' to quit."
		return s

	case stepEnterConnectionName:
		s := promptStyle.Render("Enter a name for the database connection:")
		s += "\n\n"
		if m.err != nil {
			s += fmt.Sprintf("Error: %v", m.err)
			s += "\n\n"
		}
		s += m.input.View()
		s += "\n\nPress 'Enter' to save or 'q', 'esc' to quit."
		return s

	case stepCreateSpinner:
		if m.err != nil {
			return fmt.Sprintf("❌ Error: %v\n\nPress 'q', 'esc', or Ctrl+C to quit.", m.err)
		}
		return fmt.Sprintf("Saving connection... %s", m.spinner.View())

	case stepCreateDone:
		if m.err != nil {
			return fmt.Sprintf("❌ Error: %v\n\nPress 'q', 'esc', or Ctrl+C to quit.", m.err)
		}

		return "✅ Database connection created!\n\nPress 'Enter' to continue or 'q', 'esc' to quit."

	default:
		return ""
	}
}
