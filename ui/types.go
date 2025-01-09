package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/common"
)

const (
	iconCheck       = "✅"
	iconError       = "❌"
	iconLoading     = "🔄"
	iconPrompt      = "📝"
	iconBrain       = "🤖"
	iconRun         = "🏃"
	pressToQuit     = "Press 'q' or 'esc' to quit."
	pressToQuitFull = "Press 'q', 'esc', or Ctrl+C to quit, or enter a new prompt."
)

var (
	historyEntryStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				Margin(1, 0)

	historyLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	historyNumberStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("213"))
)

type answerMsg struct {
	answer string
}

type TransitionToShellMsg struct {
	Connection common.Connection
}

type TransitionToCreateMsg struct {
	ConnectionType string
}
