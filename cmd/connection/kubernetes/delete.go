package kubernetes

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
		Short: "Delete a kubernetes connection or all kubernetes connections",
		Long:  "Delete a kubernetes connection or all kubernetes connections",
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if all {
				deleteAllKubernetesConnections()
				return
			} else if len(args) == 1 {
				deleteKubernetesConnection(args[0])
				return
			} else {
				// Interactive delete using Bubble Tea
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					deleteKubernetesConnection(selectedConnection)
				}
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all kubernetes connections")

	return deleteCmd
}

// deleteAllKubernetesConnections deletes all kubernetes connections
func deleteAllKubernetesConnections() {
	color.Yellow("Deleting all kubernetes connections")
	if err := config.DeleteAllConnections(); err != nil {
		color.Red("Error deleting all kubernetes connections: %v", err)
		return
	}
	color.Green("Deleted all kubernetes connections")
}

// deleteKubernetesConnection deletes a single kubernetes connection by name
func deleteKubernetesConnection(name string) {
	color.Yellow("Deleting kubernetes connection %s", name)
	if err := config.DeleteConnectionByName(name); err != nil {
		color.Red("Error deleting kubernetes connection: %v", err)
		return
	}
	color.Green("Deleted kubernetes connection %s", name)
}

// runInteractiveDelete runs the Bubble Tea program for interactive deletion
func runInteractiveDelete() (string, error) {
	connections, err := config.GetConnectionsByType("kubernetes")
	if err != nil {
		return "", fmt.Errorf("getting connections: %w", err)
	}

	items := make([]table.Row, len(connections))
	for i, conn := range connections {
		items[i] = table.Row{conn.Name, conn.Type, conn.SubType}
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
