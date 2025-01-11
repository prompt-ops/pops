package ui

import "github.com/prompt-ops/pops/pkg/connection"

type answerMsg struct {
	answer string
}

type TransitionToShellMsg struct {
	Connection connection.Connection
}

type TransitionToCreateMsg struct {
	ConnectionType string
}
