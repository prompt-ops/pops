package cloud

import (
	"os"

	"github.com/prompt-ops/pops/pkg/connection"
	"github.com/prompt-ops/pops/pkg/ui"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newTypesCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "types",
		Short: "List all available cloud connection types",
		Long:  "List all available cloud connection types",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListAvaibleCloudTypes(); err != nil {
				color.Red("Error listing cloud connections: %v", err)
				os.Exit(1)
			}
		},
	}

	return listCmd
}

// runListAvaibleCloudTypes lists all available cloud connection types
func runListAvaibleCloudTypes() error {
	cloudConnectionTypes := connection.AvailableCloudConnectionTypes

	items := make([]table.Row, len(cloudConnectionTypes))
	for i, cloudConnectionType := range cloudConnectionTypes {
		items[i] = table.Row{cloudConnectionType.Subtype}
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

	openTableModel := ui.NewTableModel(t, nil)

	p := tea.NewProgram(openTableModel)
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	return nil
}
