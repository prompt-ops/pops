package k8s

import (
	"fmt"

	config "github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
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
		Short: "Delete a kubernetes connection or all kubernetes connections",
		Example: `
- **pops connection kubernetes delete my-k8s-connection**: Delete a kubernetes connection named 'my-k8s-connection'.
- **pops connection kubernetes delete --all**: Delete all kubernetes connections.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if all {
				err := ui.RunWithSpinner("Deleting all kubernetes connections...", deleteAllKubernetesConnections)
				if err != nil {
					color.Red("Failed to delete all kubernetes connections: %v", err)
				}
				return
			} else if len(args) == 1 {
				connectionName := args[0]
				err := ui.RunWithSpinner(fmt.Sprintf("Deleting kubernetes connection '%s'...", connectionName), func() error {
					return deleteKubernetesConnection(connectionName)
				})
				if err != nil {
					color.Red("Failed to delete kubernetes connection '%s': %v", connectionName, err)
				}
				return
			} else {
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					err := ui.RunWithSpinner(fmt.Sprintf("Deleting kubernetes connection '%s'...", selectedConnection), func() error {
						return deleteKubernetesConnection(selectedConnection)
					})
					if err != nil {
						color.Red("Failed to delete kubernetes connection '%s': %v", selectedConnection, err)
					}
				}
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all kubernetes connections")

	return deleteCmd
}

// deleteAllKubernetesConnections deletes all kubernetes connections
func deleteAllKubernetesConnections() error {
	if err := config.DeleteAllConnectionsByType(conn.ConnectionTypeCloud); err != nil {
		return fmt.Errorf("error deleting all kubernetes connections: %w", err)
	}
	return nil
}

// deleteKubernetesConnection deletes a single kubernetes connection by name
func deleteKubernetesConnection(name string) error {
	if err := config.DeleteConnectionByName(name); err != nil {
		return fmt.Errorf("error deleting kubernetes connection: %w", err)
	}
	return nil
}

// runInteractiveDelete runs the Bubble Tea program for interactive deletion
func runInteractiveDelete() (string, error) {
	connections, err := config.GetConnectionsByType(conn.ConnectionTypeCloud)
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
