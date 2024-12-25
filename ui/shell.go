package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	config "github.com/prompt-ops/cli/config"
	connection "github.com/prompt-ops/cli/connection"
)

const (
	stepInitialChecks = iota
	stepShowContext   // New step for displaying context
	stepEnterPrompt
	stepGenerateCommand
	stepConfirmRun
	stepRunCommand
	stepDone
)

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	promptStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	commandStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	confirmationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	outputStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// historyEntry stores a single cycle of user prompt and output
type historyEntry struct {
	prompt string
	cmd    string
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
	connection     config.Connection
	popsConnection connection.PromptOpsConnection
	spinner        spinner.Model
	checkPassed    bool
}

func NewShellModel(conn config.Connection) shellModel {
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
	}
}

func (m shellModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runInitialChecks)
}

func (m shellModel) runInitialChecks() tea.Msg {
	err := m.popsConnection.CheckAuthentication()
	if err != nil {
		return errMsg{err}
	}

	err = m.popsConnection.InitialContext()
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
	case checkPassedMsg:
		m.checkPassed = true
		m.step = stepShowContext
		// Call PrintContext and set the output
		output := m.popsConnection.PrintContext()
		m.output = output
		// Proceed to the next step after displaying context
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
			switch msg.String() {
			case "up":
				if m.historyIndex > 0 && len(m.history) > 0 {
					m.historyIndex--
					previousPrompt := m.history[m.historyIndex].prompt
					m.promptInput.SetValue(previousPrompt)
					m.promptInput.CursorEnd()
				}
			case "down":
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
			case "enter":
				prompt := strings.TrimSpace(m.promptInput.Value())
				if prompt != "" {
					m.step = stepGenerateCommand
					return m, m.generateCommand(prompt)
				}
			case "q", "esc", "ctrl+c":
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
			return m, nil
		}

		// Append the successful prompt to history
		m.history = append(m.history, historyEntry{
			prompt: m.promptInput.Value(),
			cmd:    m.command,
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
		return m, tea.Quit
	}
}

func (m shellModel) View() string {
	historyView := m.renderHistory() + "\n"

	switch m.step {
	case stepInitialChecks:
		if m.checkPassed {
			return historyView + outputStyle.Render("✅ Authentication passed!\n\n")
		}
		return historyView + fmt.Sprintf(
			"%s %s",
			titleStyle.Render("🔄 Checking authentication..."),
			m.spinner.View(),
		)
	case stepShowContext:
		// This step is now handled in the Update method
		return historyView + outputStyle.Render("📄 Displaying Azure Context...\n\n"+m.output+"\n")

	case stepEnterPrompt:
		return historyView + fmt.Sprintf(
			"%s\n\n%s",
			titleStyle.Render("📝 Enter your prompt:"),
			promptStyle.Render(m.promptInput.View()),
		)
	case stepGenerateCommand:
		return historyView + titleStyle.Render("🤖 Generating command...")
	case stepConfirmRun:
		return historyView + fmt.Sprintf(
			"%s\n%s\n%s",
			commandStyle.Render(fmt.Sprintf("💡 Command: %s", m.command)),
			confirmationStyle.Render("Run this command? (Yes/No)"),
			m.confirmInput.View(),
		)
	case stepRunCommand:
		return historyView + titleStyle.Render("🏃 Running command...")
	case stepDone:
		if m.err != nil {
			return historyView + errorStyle.Render(fmt.Sprintf("❌ Error: %v\nPress 'q' or 'esc' to quit.\n", m.err))
		}
		return historyView + outputStyle.Render(
			fmt.Sprintf("✅ Output:\n%s\nPress 'q' or 'esc' or Ctrl+C to quit, or enter a new prompt.\n",
				m.output,
			),
		)
	default:
		return historyView
	}
}

// renderHistory builds a string showing all previous prompts/outputs
func (m shellModel) renderHistory() string {
	if len(m.history) == 0 {
		return ""
	}
	var out string
	for i, h := range m.history {
		out += fmt.Sprintf(
			"%d) Prompt: %s\n   Command: %s\n   Output:\n%s\n",
			i+1,
			promptStyle.Render(h.prompt),
			commandStyle.Render(h.cmd),
			outputStyle.Render(h.output),
		)
	}
	return out
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
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			return outputMsg{output: fmt.Sprintf("Error: %v", err)}
		}

		tableStr := formatAsTable(out)
		return outputMsg{output: tableStr}
	}
}

func formatAsTable(output []byte) string {
	lines := splitLines(output)
	if len(lines) == 0 {
		return string(output) // fallback if no lines
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	// Parse the first line as headers
	headers := strings.Fields(lines[0])
	table.SetHeader(headers)
	headerCount := len(headers)

	// Parse the remaining lines as rows
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		// If we have more fields than headers, merge the extras into the last column
		if len(fields) > headerCount {
			merged := strings.Join(fields[headerCount-1:], " ")
			fields = append(fields[:headerCount-1], merged)
		}
		table.Append(fields)
	}

	table.Render()
	return buf.String()
}

func splitLines(output []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var lines []string
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			lines = append(lines, text)
		}
	}
	return lines
}

type commandMsg struct {
	command string
}

type outputMsg struct {
	output string
}
