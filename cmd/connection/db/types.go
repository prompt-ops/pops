package db

import (
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/prompt-ops/pops/connection/db"
	commonui "github.com/prompt-ops/pops/ui/common"
	"github.com/spf13/cobra"
)

func newTypesCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "types",
		Short: "List all available database connection types",
		Long:  "List all available database connection types",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListAvaibleDatabaseTypes(); err != nil {
				color.Red("Error listing database connections: %v", err)
				os.Exit(1)
			}
		},
	}

	return listCmd
}

// runListAvaibledatabaseTypes lists all available database connection types
func runListAvaibleDatabaseTypes() error {
	connectionTypes := db.AvailableConnectionTypes()

	items := make([]table.Row, len(connectionTypes))
	for i, connectionType := range connectionTypes {
		items[i] = table.Row{connectionType}
	}

	columns := []table.Column{
		{Title: "Available Types", Width: 25},
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
