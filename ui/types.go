package ui

import "github.com/prompt-ops/pops/common"

type TransitionToShellMsg struct {
	Connection common.Connection
}

type TransitionToCreateMsg struct {
	ConnectionType string
}
