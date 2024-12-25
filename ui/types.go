package ui

import config "github.com/prompt-ops/cli/config"

// Sent when it’s time to transition to the shell
type TransitionToShellMsg struct {
	Connection config.Connection
}
