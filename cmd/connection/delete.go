package connection

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	config "github.com/prompt-ops/pops/config"
	commonui "github.com/prompt-ops/pops/ui/common"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the delete command
func newDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a connection or all connections",
		Long:  "Delete a connection or all connections",
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if all {
				deleteAllConnections()
				return
			} else if len(args) == 1 {
				deleteConnection(args[0])
				return
			} else {
				// Interactive delete using Bubble Tea
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					deleteConnection(selectedConnection)
				}
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all connections")

	return deleteCmd
}

// deleteAllConnections deletes all connections
func deleteAllConnections() {
	color.Yellow("Deleting all connections")
	if err := config.DeleteAllConnections(); err != nil {
		color.Red("Error deleting all connections: %v", err)
		return
	}
	color.Green("Deleted all connections")
}

// deleteConnection deletes a single connection by name
func deleteConnection(name string) {
	color.Yellow("Deleting connection %s", name)
	if err := config.DeleteConnectionByName(name); err != nil {
		color.Red("Error deleting connection: %v", err)
		return
	}
	color.Green("Deleted connection %s", name)
}

// runInteractiveDelete runs the Bubble Tea program for interactive deletion
func runInteractiveDelete() (string, error) {
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

	deleteTableModel := commonui.NewTableModel(t, nil)

	p := tea.NewProgram(deleteTableModel)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("running Bubble Tea program: %w", err)
	}

	return deleteTableModel.Selected(), nil
}
