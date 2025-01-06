package connection

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prompt-ops/pops/cmd/connection/factory"
	"github.com/prompt-ops/pops/common"
	"github.com/prompt-ops/pops/ui"
	commonui "github.com/prompt-ops/pops/ui/common"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new connection",
		Long:    "Create a new connection",
		Example: `pops connection create`,
		Run: func(cmd *cobra.Command, args []string) {
			// `pops connection create` interactive command.
			// This command will be used to create a new connection.
			runInteractiveCreate()
		},
	}

	return createCmd
}

func runInteractiveCreate() {
	m := initialCreateModel()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

type createConnectionStep int

const (
	createStepTypeSelection createConnectionStep = iota
	createStepCreateModel
)

type createModel struct {
	currentStep        createConnectionStep
	typeSelectionModel tea.Model
	createModel        tea.Model
}

func initialCreateModel() *createModel {
	connectionTypes := common.AvailableConnectionTypes()

	items := make([]table.Row, len(connectionTypes))
	for i, connectionType := range connectionTypes {
		items[i] = table.Row{
			connectionType,
		}
	}

	columns := []table.Column{
		{Title: "Type", Width: 25},
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

	onSelect := func(selectedType string) tea.Msg {
		return ui.TransitionToCreateMsg{
			ConnectionType: selectedType,
		}
	}

	typeSelectionModel := commonui.NewTableModel(t, onSelect)

	return &createModel{
		currentStep:        createStepTypeSelection,
		typeSelectionModel: typeSelectionModel,
	}
}

func (m *createModel) Init() tea.Cmd {
	return m.typeSelectionModel.Init()
}

func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case createStepTypeSelection:
		switch msg := msg.(type) {
		case ui.TransitionToCreateMsg:
			connectionType := msg.ConnectionType
			createModel, err := factory.GetCreateModel(connectionType)
			if err != nil {
				return m, tea.Quit
			}

			m.currentStep = createStepCreateModel
			m.createModel = createModel
			return m, createModel.Init()
		default:
			var cmd tea.Cmd
			m.typeSelectionModel, cmd = m.typeSelectionModel.Update(msg)
			return m, cmd
		}
	case createStepCreateModel:
		var cmd tea.Cmd
		m.createModel, cmd = m.createModel.Update(msg)
		return m, cmd
	default:
		return m, tea.Quit
	}
}

func (m *createModel) View() string {
	switch m.currentStep {
	case createStepTypeSelection:
		return m.typeSelectionModel.View()
	case createStepCreateModel:
		return m.createModel.View()
	default:
		return "Unknown step"
	}
}
