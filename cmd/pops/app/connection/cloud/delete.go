package cloud

import (
	"fmt"

	"github.com/prompt-ops/pops/pkg/config"
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
	var name string

	deleteCmd := &cobra.Command{
		Use:   "delete [connection-name]",
		Short: "Delete a cloud connection or all cloud connections",
		Long: `Delete a cloud connection or all cloud connections.

You can specify the connection name either as a positional argument or using the --name flag.

Examples:
  pops connection cloud delete my-cloud-connection
  pops connection cloud delete --name my-cloud-connection
  pops connection cloud delete --all
`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if all {
				// If --all flag is provided, ignore other arguments and flags
				err := ui.RunWithSpinner("Deleting all cloud connections...", deleteAllCloudConnections)
				if err != nil {
					color.Red("Failed to delete all cloud connections: %v", err)
				}
				return
			}

			var connectionName string

			// Determine the connection name based on --name flag and positional arguments
			if name != "" && len(args) > 0 {
				// If both --name flag and positional argument are provided, prioritize the flag
				fmt.Println("Warning: --name flag is provided; ignoring positional argument.")
				connectionName = name
			} else if name != "" {
				// If only --name flag is provided
				connectionName = name
			} else if len(args) == 1 {
				// If only positional argument is provided
				connectionName = args[0]
			} else {
				// Interactive mode if neither --name flag nor positional argument is provided
				selectedConnection, err := runInteractiveDelete()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}
				if selectedConnection != "" {
					err := ui.RunWithSpinner(fmt.Sprintf("Deleting cloud connection '%s'...", selectedConnection), func() error {
						return deleteCloudConnection(selectedConnection)
					})
					if err != nil {
						color.Red("Failed to delete cloud connection '%s': %v", selectedConnection, err)
					}
				}
				return
			}

			// Non-interactive mode: Delete the specified connection
			err = ui.RunWithSpinner(fmt.Sprintf("Deleting cloud connection '%s'...", connectionName), func() error {
				return deleteCloudConnection(connectionName)
			})
			if err != nil {
				color.Red("Failed to delete cloud connection '%s': %v", connectionName, err)
			}
		},
	}

	// Define the --name flag
	deleteCmd.Flags().StringVar(&name, "name", "", "Name of the cloud connection to delete")
	// Define the --all flag
	deleteCmd.Flags().Bool("all", false, "Delete all cloud connections")

	return deleteCmd
}

// deleteAllCloudConnections deletes all cloud connections
func deleteAllCloudConnections() error {
	if err := config.DeleteAllConnectionsByType(connection.ConnectionTypeCloud); err != nil {
		return fmt.Errorf("error deleting all cloud connections: %w", err)
	}
	color.Green("All cloud connections have been successfully deleted.")
	return nil
}

// deleteCloudConnection deletes a single cloud connection by name
func deleteCloudConnection(name string) error {
	// Check if the connection exists before attempting to delete
	conn, err := getConnectionByName(name)
	if err != nil {
		return fmt.Errorf("connection '%s' does not exist", name)
	}

	if err := config.DeleteConnectionByName(name); err != nil {
		return fmt.Errorf("error deleting cloud connection: %w", err)
	}

	color.Green("Cloud connection '%s' has been successfully deleted.", conn.Name)
	return nil
}

// runInteractiveDelete runs the Bubble Tea program for interactive deletion
func runInteractiveDelete() (string, error) {
	connections, err := config.GetConnectionsByType(connection.ConnectionTypeCloud)
	if err != nil {
		return "", fmt.Errorf("getting connections: %w", err)
	}

	if len(connections) == 0 {
		return "", fmt.Errorf("no cloud connections available to delete")
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
