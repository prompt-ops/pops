package models

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type endModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
	done     bool
}

func initialEndModel() endModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return endModel{spinner: s}
}

func (m endModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runTask)
}

func (m endModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case taskCompleteMsg:
		m.done = true
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m endModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	if m.done {
		return "\n\n   Task completed successfully!\n\n"
	}
	str := fmt.Sprintf("\n\n   %s Running task...press q to quit\n\n", m.spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
}

func (m endModel) Quitting() bool {
	return m.quitting
}

func (m endModel) Done() bool {
	return m.done
}

func NewEndModel() endModel {
	return initialEndModel()
}

type taskCompleteMsg struct{}

func runTask() tea.Msg {
	// Simulate a long-running task
	time.Sleep(5 * time.Second)
	return taskCompleteMsg{}
}
