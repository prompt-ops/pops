package cloud

import (
	"fmt"
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
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	outputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

const (
	stepSelectProvider step = iota
	stepEnterConnectionName
	stepCreateSpinner
	stepCreateDone
)

var providers = conn.AvailableCloudConnectionTypes

type (
	doneWaitingMsg struct {
		Connection conn.Connection
	}

	errMsg struct {
		err error
	}
)

type createModel struct {
	currentStep step
	cursor      int
	input       textinput.Model
	err         error
	spinner     spinner.Model

	connection            conn.Connection
	selectedCloudProvider conn.AvailableCloudConnectionType
}

func NewCreateModel() *createModel {
	ti := textinput.New()
	ti.Placeholder = ui.EnterConnectionNameMessage
	ti.CharLimit = 256
	ti.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &createModel{
		currentStep: stepSelectProvider,
		input:       ti,
		spinner:     sp,
	}
}

func (m *createModel) Init() tea.Cmd {
	// Make the text input blink by default
	return textinput.Blink
}

func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentStep {

	//----------------------------------------------------------------------
	// stepSelectProvider
	//----------------------------------------------------------------------
	case stepSelectProvider:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(providers)-1 {
					m.cursor++
				}
			case "enter":
				m.selectedCloudProvider = providers[m.cursor]
				m.currentStep = stepEnterConnectionName
				m.input.Focus()
				m.err = nil
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}

	//----------------------------------------------------------------------
	// stepEnterConnectionName
	//----------------------------------------------------------------------
	case stepEnterConnectionName:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m.input, cmd = m.input.Update(msg)
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

				connection := conn.NewCloudConnection(name, m.selectedCloudProvider)
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
		default:
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	//----------------------------------------------------------------------
	// stepCreateSpinner
	//----------------------------------------------------------------------
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

	//----------------------------------------------------------------------
	// stepCreateDone
	//----------------------------------------------------------------------
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
		}
	}

	return m, cmd
}

func waitTwoSecondsCmd(conn conn.Connection) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return doneWaitingMsg{
			Connection: conn,
		}
	})
}

func (m *createModel) View() string {
	// Clear the terminal before rendering the UI
	clearScreen := "\033[H\033[2J"

	switch m.currentStep {
	case stepSelectProvider:
		title := "Select a cloud provider (↑/↓, Enter to confirm):"
		footer := ui.QuitMessage

		var subtypeSelection string
		for i, p := range providers {
			cursor := "  "
			if i == m.cursor {
				cursor = "→ "
			}
			subtypeSelection += fmt.Sprintf("%s%s\n", cursor, promptStyle.Render(p.Subtype))
		}

		return fmt.Sprintf(
			"%s\n\n%s\n%s",
			titleStyle.Render(title),
			subtypeSelection,
			lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(footer),
		)

	case stepEnterConnectionName:
		title := "Enter a name for the Cloud connection:"
		footer := ui.QuitMessage

		if m.err != nil {
			errorMessage := fmt.Sprintf("Error: %v", m.err)

			return fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render(title),
				errorStyle.Render(errorMessage),
				promptStyle.Render(m.input.View()),
				lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(footer),
			)
		}

		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render(title),
			promptStyle.Render(m.input.View()),
			lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(footer),
		)

	case stepCreateSpinner:
		return clearScreen + outputStyle.Render("Saving connection... ") + m.spinner.View()

	case stepCreateDone:
		if m.err != nil {
			return clearScreen + errorStyle.Render(fmt.Sprintf("❌ Error: %v\n\nPress 'Enter' or 'q'/'esc' to quit.", m.err))
		}

		return clearScreen + outputStyle.Render(
			"✅ Cloud connection created!\n\nPress 'Enter' or 'q'/'esc' to exit.",
		)

	default:
		return clearScreen
	}
}
