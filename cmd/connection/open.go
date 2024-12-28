package connection

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/prompt-ops/pops/common"
	"github.com/prompt-ops/pops/config"
	"github.com/prompt-ops/pops/ui"
	commonui "github.com/prompt-ops/pops/ui/common"
	"github.com/spf13/cobra"
)

// newOpenCmd creates the open command
func newOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open",
		Short: "Open a connection",
		Long:  "Open a connection",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				// Check if connection exists
				conn, err := config.GetConnectionByName(args[0])
				if err != nil {
					color.Red("Error getting connection: %v", err)
					return
				}
				// Open the shell UI
				if err := openShellUI(conn); err != nil {
					color.Red("Error opening shell UI: %v", err)
				}
			} else {
				// Interactive open using Bubble Tea
				selectedConnection, err := runInteractiveOpen()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					// Check if connection exists
					conn, err := config.GetConnectionByName(selectedConnection)
					if err != nil {
						color.Red("Error getting connection: %v", err)
						return
					}
					// Open the shell UI
					if err := openShellUI(conn); err != nil {
						color.Red("Error opening shell UI: %v", err)
					}
				}
			}
		},
	}

	return openCmd
}

// runInteractiveOpen runs the Bubble Tea program for interactive open
func runInteractiveOpen() (string, error) {
	connections, err := config.GetAllConnections()
	if err != nil {
		return "", fmt.Errorf("getting connections: %w", err)
	}

	items := make([]table.Row, len(connections))
	for i, conn := range connections {
		items[i] = table.Row{conn.Name, conn.Type.GetMainType(), conn.Type.GetSubtype()}
	}

	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 15},
		{Title: "Subtype", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(items),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	openTableModel := commonui.NewTableModel(t, nil)

	p := tea.NewProgram(openTableModel)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("running Bubble Tea program: %w", err)
	}

	return openTableModel.Selected(), nil
}

// openShellUI initializes and runs the shell UI with the selected connection
func openShellUI(conn common.Connection) error {
	shell := ui.NewShellModel(conn)
	p := tea.NewProgram(shell)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start shell UI: %w", err)
	}
	return nil
}
