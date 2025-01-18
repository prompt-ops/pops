package ui

import (
	"github.com/prompt-ops/pops/pkg/conn"
)

type answerMsg struct {
	answer string
}

type TransitionToShellMsg struct {
	Connection conn.Connection
}

type TransitionToCreateMsg struct {
	ConnectionType string
}

var (
	// EnterConnectionNameMessage is the message displayed when the user is prompted to enter a connection name.
	EnterConnectionNameMessage = "Enter connection name:"

	// QuitMessage is the message displayed to show user how to quit the application.
	QuitMessage = "Press 'q' or 'esc' or Ctrl+C to quit."
)
