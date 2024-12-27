package kubernetes

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	config "github.com/prompt-ops/pops/config"
	commonui "github.com/prompt-ops/pops/ui/common"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all kubernetes connections",
		Long:  "List all kubernetes connections that have been set up.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListConnections(); err != nil {
				color.Red("Error listing kubernetes connections: %v", err)
				os.Exit(1)
			}
		},
	}

	return listCmd
}

// runListConnections lists all connections
func runListConnections() error {
	connections, err := config.GetConnectionsByType("kubernetes")
	if err != nil {
		return fmt.Errorf("getting kubernetes connections: %w", err)
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

	openTableModel := commonui.NewTableModel(t, nil)

	p := tea.NewProgram(openTableModel)
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	return nil
}