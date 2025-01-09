package ui

import (
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

type answerMsg struct {
	answer string
}

type TransitionToShellMsg struct {
	Connection common.Connection
}

type TransitionToCreateMsg struct {
	ConnectionType string
}
