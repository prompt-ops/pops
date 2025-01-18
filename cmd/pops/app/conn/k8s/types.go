package k8s

import (
	"os"

	"github.com/prompt-ops/pops/pkg/conn"
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
		Short: "List all available kubernetes connection types",
		Long:  "List all available kubernetes connection types",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListAvaibleKubernetesTypes(); err != nil {
				color.Red("Error listing kubernetes connections: %v", err)
				os.Exit(1)
			}
		},
	}

	return listCmd
}

// runListAvaibleKubernetesTypes lists all available kubernetes connection types
func runListAvaibleKubernetesTypes() error {
	connectionTypes := conn.AvailableKubernetesConnectionTypes

	items := make([]table.Row, len(connectionTypes))
	for i, connectionType := range connectionTypes {
		items[i] = table.Row{connectionType.Subtype}
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
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("212")).
		Bold(true)
	t.SetStyles(s)

	openTableModel := ui.NewTableModel(t, nil, true)

	p := tea.NewProgram(openTableModel)
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	return nil
}
