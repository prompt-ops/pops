package db

import (
	"fmt"

	config "github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/connection"
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
		Short: "Delete a database connection or all database connections",
		Example: `
- **pops connection db delete my-db-connection**: Delete a database connection named 'my-db-connection'.
- **pops connection db delete --all**: Delete all database connections.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if all {
				err := ui.RunWithSpinner("Deleting all database connections...", deleteAllDatabaseConnections)
				if err != nil {
					color.Red("Failed to delete all database connections: %v", err)
				}
				return
			} else if len(args) == 1 {
				connectionName := args[0]
				err := ui.RunWithSpinner(fmt.Sprintf("Deleting database connection '%s'...", connectionName), func() error {
					return deleteDatabaseConnection(connectionName)
				})
				if err != nil {
					color.Red("Failed to delete database connection '%s': %v", connectionName, err)
				}
				return
			} else {
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					err := ui.RunWithSpinner(fmt.Sprintf("Deleting database connection '%s'...", selectedConnection), func() error {
						return deleteDatabaseConnection(selectedConnection)
					})
					if err != nil {
						color.Red("Failed to delete database connection '%s': %v", selectedConnection, err)
					}
				}
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all database connections")

	return deleteCmd
}

// deleteAllDatabaseConnections deletes all database connections
func deleteAllDatabaseConnections() error {
	if err := config.DeleteAllConnectionsByType(connection.ConnectionTypeDatabase); err != nil {
		return fmt.Errorf("error deleting all database connections: %w", err)
	}
	return nil
}

// deleteDatabaseConnection deletes a single database connection by name
func deleteDatabaseConnection(name string) error {
	if err := config.DeleteConnectionByName(name); err != nil {
		return fmt.Errorf("error deleting database connection: %w", err)
	}
	return nil
}

// runInteractiveDelete runs the Bubble Tea program for interactive deletion
func runInteractiveDelete() (string, error) {
	connections, err := config.GetConnectionsByType(connection.ConnectionTypeDatabase)
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
		{Title: "Driver", Width: 20},
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

	deleteTableModel := ui.NewTableModel(t, nil)

	p := tea.NewProgram(deleteTableModel)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("running Bubble Tea program: %w", err)
	}

	return deleteTableModel.Selected(), nil
}
