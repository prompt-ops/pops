package connection

import (
	"fmt"

	config "github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/ui"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
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
				err := ui.RunWithSpinner("Deleting all connections...", deleteAllConnections)
				if err != nil {
					color.Red("Failed to delete all connections: %v", err)
				}
				return
			} else if len(args) == 1 {
				connectionName := args[0]
				err := ui.RunWithSpinner(fmt.Sprintf("Deleting connection '%s'...", connectionName), func() error {
					return deleteConnection(connectionName)
				})
				if err != nil {
					color.Red("Failed to delete connection '%s': %v", connectionName, err)
				}
				return
			} else {
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					err := ui.RunWithSpinner(fmt.Sprintf("Deleting connection '%s'...", selectedConnection), func() error {
						return deleteConnection(selectedConnection)
					})
					if err != nil {
						color.Red("Failed to delete connection '%s': %v", selectedConnection, err)
					}
				}
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all connections")

	return deleteCmd
}

// deleteAllConnections deletes all connections
func deleteAllConnections() error {
	if err := config.DeleteAllConnections(); err != nil {
		return fmt.Errorf("error deleting all connections: %w", err)
	}
	return nil
}

// deleteConnection deletes a single connection by name
func deleteConnection(name string) error {
	if err := config.DeleteConnectionByName(name); err != nil {
		return fmt.Errorf("error deleting connection '%s': %w", name, err)
	}
	return nil
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
		{Title: "Name", Width: 15},
		{Title: "Type", Width: 15},
		{Title: "Subtype", Width: 15},
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
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("212")).
		Bold(true)
	t.SetStyles(s)

	deleteTableModel := ui.NewTableModel(t, nil, false)

	p := tea.NewProgram(deleteTableModel)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("running Bubble Tea program: %w", err)
	}

	return deleteTableModel.Selected(), nil
}
