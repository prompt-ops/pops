package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/pkg/connection"
)

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	promptStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	commandStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	confirmationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	outputStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

var (
	historyContainerStyle = lipgloss.NewStyle().
				Width(72).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				Margin(1, 0)

	historyLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))
)

type queryMode int

const (
	modeCommand queryMode = iota
	modeAnswer
)

const (
	stepInitialChecks = iota
	stepShowContext
	stepEnterPrompt
	stepGenerateCommand
	stepGetAnswer
	stepConfirmRun
	stepRunCommand
	stepDone
)

// historyEntry stores a single cycle of user prompt and output
type historyEntry struct {
	prompt string
	cmd    string

	// Command or Answer
	mode string

	output string
	err    error
}

type shellModel struct {
	step           int
	promptInput    textinput.Model
	command        string
	confirmInput   textinput.Model
	output         string
	err            error
	history        []historyEntry
	historyIndex   int
	connection     connection.Connection
	popsConnection connection.ConnectionInterface
	spinner        spinner.Model
	checkPassed    bool
	mode           queryMode
	windowWidth    int
}

func NewShellModel(conn connection.Connection) shellModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your prompt..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	ci := textinput.New()
	ci.Placeholder = "Yes/No"
	ci.CharLimit = 3
	ci.Width = 10

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Get the right connection implementation
	popsConn, err := connection.GetConnection(conn)
	if err != nil {
		panic(err)
	}

	return shellModel{
		step:           stepInitialChecks,
		promptInput:    ti,
		confirmInput:   ci,
		history:        []historyEntry{},
		connection:     conn,
		popsConnection: popsConn,
		spinner:        sp,
		mode:           modeCommand,
	}
}

func (m shellModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runInitialChecks, tea.EnterAltScreen)
}

func (m shellModel) runInitialChecks() tea.Msg {
	err := m.popsConnection.CheckAuthentication()
	if err != nil {
		return errMsg{err}
	}

	err = m.popsConnection.SetContext()
	if err != nil {
		return errMsg{err}
	}

	// Add other initial checks here if needed
	return checkPassedMsg{}
}

type checkPassedMsg struct{}
type errMsg struct{ err error }

func (m shellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil

	case checkPassedMsg:
		m.checkPassed = true
		m.step = stepShowContext
		m.output = "Will be added here"
		m.step = stepEnterPrompt
		return m, textinput.Blink

	case errMsg:
		m.err = msg.err
		m.step = stepDone
		return m, nil
	}

	switch m.step {
	case stepInitialChecks:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepShowContext:
		// This step is now handled in the checkPassedMsg case
		return m, nil

	case stepEnterPrompt:
		var cmd tea.Cmd
		m.promptInput, cmd = m.promptInput.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyUp:
				if m.historyIndex > 0 && len(m.history) > 0 {
					m.historyIndex--
					previousPrompt := m.history[m.historyIndex].prompt
					m.promptInput.SetValue(previousPrompt)
					m.promptInput.CursorEnd()
				}
			case tea.KeyDown:
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					nextPrompt := m.history[m.historyIndex].prompt
					m.promptInput.SetValue(nextPrompt)
					m.promptInput.CursorEnd()
				} else {
					// Clear the input if at the latest entry
					m.historyIndex = len(m.history)
					m.promptInput.SetValue("")
				}

			case tea.KeyLeft, tea.KeyRight:
				// Toggle between Command and Answer mode
				if m.mode == modeCommand {
					m.mode = modeAnswer
				} else {
					m.mode = modeCommand
				}

			case tea.KeyEnter:
				prompt := strings.TrimSpace(m.promptInput.Value())
				if prompt != "" {
					if m.mode == modeCommand {
						m.step = stepGenerateCommand
						return m, m.generateCommand(prompt)
					} else {
						m.step = stepGetAnswer
						return m, m.generateAnswer(prompt)
					}
				}

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		}
		return m, cmd

	case stepGenerateCommand:
		if cmdMsg, ok := msg.(commandMsg); ok {
			m.command = cmdMsg.command
			m.step = stepConfirmRun
			m.confirmInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case stepGetAnswer:
		// If we successfully got an answer, we’ll receive an `answerMsg`.
		if ansMsg, ok := msg.(answerMsg); ok {
			m.output = ansMsg.answer
			m.step = stepDone
			return m, nil
		}
		return m, nil

	case stepConfirmRun:
		var cmd tea.Cmd
		m.confirmInput, cmd = m.confirmInput.Update(msg)
		if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
			val := m.confirmInput.Value()
			if val == "Yes" || val == "yes" {
				m.step = stepRunCommand
				return m, m.runCommand(m.command)
			} else if val == "No" || val == "no" {
				// User declined, reset prompt
				m.step = stepEnterPrompt
				m.promptInput.Reset()
				m.confirmInput.Reset()
				m.historyIndex = len(m.history) // Reset historyIndex
				return m, textinput.Blink
			}
		}
		return m, cmd

	case stepRunCommand:
		if outMsg, ok := msg.(outputMsg); ok {
			m.output = outMsg.output
			m.step = stepDone
			return m, nil
		}
		return m, nil

	case stepDone:
		if m.err != nil {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "q", "esc", "ctrl+c":
					return m, tea.Quit
				case "enter":
					// Reset the model to start a new prompt
					m.err = nil
					m.step = stepEnterPrompt
					m.promptInput.Reset()
					return m, textinput.Blink
				}
			}
			return m, nil
		}

		mode := "Command"
		if m.mode == modeAnswer {
			mode = "Answer"
		}

		// Append the successful prompt to history
		m.history = append(m.history, historyEntry{
			prompt: m.promptInput.Value(),
			cmd:    m.command,
			mode:   mode,
			output: m.output,
			err:    m.err,
		})
		// Reset historyIndex to point to the end
		m.historyIndex = len(m.history)
		// Reset inputs for next prompt
		m.step = stepEnterPrompt
		m.promptInput.Reset()
		m.confirmInput.Reset()

		return m, textinput.Blink

	default:
		// Unexpected step; exit
		return m, tea.Quit
	}
}

func (m shellModel) View() string {
	historyView := lipgloss.NewStyle().
		MaxWidth(m.windowWidth-2).
		Margin(0, 1).
		Render(m.renderHistory())

	var content string

	switch m.step {
	case stepInitialChecks:
		content = m.viewInitialChecks()

	case stepEnterPrompt:
		content = m.viewEnterPrompt()

	case stepGenerateCommand:
		content = m.viewGenerateCommand()

	case stepGetAnswer:
		content = m.viewGetAnswer()

	case stepConfirmRun:
		content = m.viewConfirmRun()

	case stepRunCommand:
		content = m.viewRunCommand()

	case stepDone:
		content = m.viewDone()

	default:
		content = ""
	}

	return lipgloss.JoinVertical(lipgloss.Top, historyView, content)
}

func (m shellModel) viewInitialChecks() string {
	if m.checkPassed {
		return outputStyle.Render("✅ Authentication passed!\n\n")
	}
	return fmt.Sprintf(
		"%s %s",
		m.spinner.View(),
		titleStyle.Render("Checking Authentication..."),
	)
}

func (m shellModel) viewEnterPrompt() string {
	var title string
	var modeStr string

	if m.mode == modeCommand {
		title = "🤖 Request a command/query"
		modeStr = "command/query"
	} else {
		title = "💡 Ask a question"
		modeStr = "answer"
	}

	footer := "Use ←/→ to switch between modes (currently " + modeStr + "). Press Enter when ready."

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		promptStyle.Render(m.promptInput.View()),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(footer),
	)
}

func (m shellModel) viewGenerateCommand() string {
	return titleStyle.Render("🤖 Generating command...")
}

func (m shellModel) viewGetAnswer() string {
	return titleStyle.Render("🤔 Getting your answer...")
}

func (m shellModel) viewConfirmRun() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		commandStyle.Render(fmt.Sprintf("💡 Command: %s", m.command)),
		confirmationStyle.Render("Run this command? (Yes/No)"),
		m.confirmInput.View(),
	)
}

func (m shellModel) viewRunCommand() string {
	return titleStyle.Render("🏃 Running command...")
}

func (m shellModel) viewDone() string {
	width := 72
	if m.windowWidth < 72 {
		width = m.windowWidth - 2
	}

	outStyle := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width)

	var content string
	if m.err != nil {
		content = fmt.Sprintf("%v\n", m.err)
		content = errorStyle.Render(content)
	} else {
		content = fmt.Sprintf("%s\n", m.output)
		content = outputStyle.Render(content)
	}

	content = outStyle.Render(content)
	footer := "Press 'q' or 'esc' or Ctrl+C to quit, or enter a new prompt."
	return lipgloss.JoinVertical(lipgloss.Top, content, footer)
}

// renderHistory builds a string showing all previous prompts/outputs
func (m shellModel) renderHistory() string {
	if len(m.history) == 0 {
		return ""
	}

	var entries []string
	for _, h := range m.history {
		// Build lines with label + content
		promptLine := lipgloss.JoinHorizontal(
			lipgloss.Top,
			historyLabelStyle.Render("Prompt: "),
			promptStyle.Render(h.prompt),
		)

		var modeLine string
		if h.mode == "Command" {
			modeLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Command: "),
				commandStyle.Render(h.cmd),
			)
		} else {
			modeLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Answer: "),
			)
		}

		outputLine := lipgloss.JoinHorizontal(
			lipgloss.Top,
			outputStyle.Render(h.output),
		)

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			promptLine,
			modeLine,
			outputLine,
		)

		content = lipgloss.JoinVertical(lipgloss.Left, content)

		// Render the box
		boxed := historyContainerStyle.
			MaxWidth(m.windowWidth - 2).
			Render(content)
		entries = append(entries, boxed)
	}

	return lipgloss.JoinVertical(lipgloss.Left, entries...)
}

func (m shellModel) generateCommand(prompt string) tea.Cmd {
	return func() tea.Msg {
		cmd, err := m.popsConnection.GetCommand(prompt)
		if err != nil {
			return errMsg{err}
		}

		return commandMsg{
			command: cmd,
		}
	}
}

func (m shellModel) runCommand(command string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.popsConnection.ExecuteCommand(command)
		if err != nil {
			return errMsg{err}
		}

		outStr, err := m.popsConnection.FormatResultAsTable(out)
		if err != nil {
			return errMsg{err}
		}

		return outputMsg{
			output: outStr,
		}
	}
}

func (m shellModel) generateAnswer(prompt string) tea.Cmd {
	return func() tea.Msg {
		answer, err := m.popsConnection.GetAnswer(prompt)
		if err != nil {
			return errMsg{err}
		}

		return answerMsg{
			answer,
		}
	}
}

type commandMsg struct {
	command string
}

type outputMsg struct {
	output string
}
