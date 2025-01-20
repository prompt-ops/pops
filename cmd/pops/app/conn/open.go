package conn

import (
	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/ui/conn"
	"github.com/prompt-ops/pops/pkg/ui/shell"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// newOpenCmd creates the open command for the connection.
func newOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open",
		Short: "Open a connection",
		Long:  "Open a connection",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				openSingleConnection(args[0])
			} else {
				openConnectionPicker()
			}
		},
	}

	return openCmd
}

// openSingleConnection opens a single connection by name.
func openSingleConnection(name string) {
	conn, err := config.GetConnectionByName(name)
	if err != nil {
		color.Red("Error getting connection: %v", err)
		return
	}

	shell := shell.NewShellModel(conn)
	p := tea.NewProgram(shell)
	if _, err := p.Run(); err != nil {
		color.Red("Error opening shell UI: %v", err)
	}
}

// openConnectionPicker opens the connection picker UI.
func openConnectionPicker() {
	root := conn.NewOpenRootModel()
	p := tea.NewProgram(root)
	if _, err := p.Run(); err != nil {
		color.Red("Error: %v", err)
	}
}
