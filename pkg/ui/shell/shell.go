package shell

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/pkg/conn"
	"golang.org/x/term"
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
	connection     conn.Connection
	popsConnection conn.ConnectionInterface
	spinner        spinner.Model
	checkPassed    bool
	mode           queryMode
	windowWidth    int
}

func NewShellModel(connection conn.Connection) shellModel {
	ti := textinput.New()
	ti.Placeholder = "Define the command or query to be generated via Prompt-Ops..."
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 100

	ci := textinput.New()
	ci.Placeholder = "Y/n"
	ci.CharLimit = 3
	ci.Width = 100
	ci.PromptStyle.Padding(0, 1)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Get the right connection implementation
	popsConn, err := conn.GetConnection(connection)
	if err != nil {
		panic(err)
	}

	return shellModel{
		step:           stepInitialChecks,
		promptInput:    ti,
		confirmInput:   ci,
		history:        []historyEntry{},
		connection:     connection,
		popsConnection: popsConn,
		spinner:        sp,
		mode:           modeCommand,
	}
}

func (m shellModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runInitialChecks,
		tea.EnterAltScreen,
		requestWindowSize(),
	)
}

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
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyF1 {
				m.step = stepEnterPrompt
			}
		}
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
					m.historyIndex = len(m.history)
					m.promptInput.SetValue("")
				}

			case tea.KeyLeft, tea.KeyRight:
				if m.mode == modeCommand {
					m.mode = modeAnswer
				} else {
					m.mode = modeCommand
				}
				m.updatePromptInputPlaceholder()

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

			case tea.KeyF1:
				m.step = stepShowContext
				output, err := m.popsConnection.GetFormattedContext()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.output = output
				return m, nil
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
			if val == "Y" || val == "y" {
				m.step = stepRunCommand
				return m, m.runCommand(m.command)
			} else if val == "N" || val == "n" {
				m.step = stepEnterPrompt
				m.promptInput.Reset()
				m.confirmInput.Reset()
				m.historyIndex = len(m.history)
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
			if key, ok := msg.(tea.KeyMsg); ok {
				switch key.String() {
				case "q", "esc", "ctrl+c":
					return m, tea.Quit
				case "enter":
					m.err = nil
					m.step = stepEnterPrompt
					m.promptInput.Reset()
					return m, textinput.Blink
				}
			}
			return m, nil
		}

		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			case "enter":
				mode := "Command"
				if m.mode == modeAnswer {
					mode = "Answer"
				}

				m.history = append(m.history, historyEntry{
					prompt: m.promptInput.Value(),
					cmd:    m.command,
					mode:   mode,
					output: m.output,
					err:    m.err,
				})

				m.historyIndex = len(m.history)
				m.step = stepEnterPrompt
				m.promptInput.Reset()
				m.confirmInput.Reset()
				return m, textinput.Blink
			}
		}

		return m, nil

	default:
		return m, tea.Quit
	}
}

func (m shellModel) View() string {
	historyView := lipgloss.NewStyle().
		MaxWidth(m.windowWidth-2).
		Margin(0, 1).
		Render(m.viewHistory())

	var content string

	switch m.step {
	case stepInitialChecks:
		content = m.viewInitialChecks()

	case stepShowContext:
		content = m.viewShowContext()

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

func requestWindowSize() tea.Cmd {
	return func() tea.Msg {
		w, h, err := term.GetSize(0)
		if err != nil {
			w, h = 80, 24
		}
		return tea.WindowSizeMsg{
			Width:  w,
			Height: h,
		}
	}
}

func (m *shellModel) updatePromptInputPlaceholder() {
	if m.mode == modeAnswer {
		m.promptInput.Placeholder = "Ask a question via Prompt-Ops..."
	} else {
		m.promptInput.Placeholder = "Define the command or query to be generated via Prompt-Ops..."
	}
}
