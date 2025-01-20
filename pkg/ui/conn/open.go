package conn

import (
	"fmt"

	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/ui"
	"github.com/prompt-ops/pops/pkg/ui/shell"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

type openRootModel struct {
	step       openStep
	tableModel tableModel
	shellModel tea.Model
}

type openStep int

const (
	stepPick openStep = iota
	stepShell
)

type tableModel interface {
	tea.Model
	Selected() string
}

// NewOpenRootModel initializes the openRootModel with connections.
func NewOpenRootModel() *openRootModel {
	connections, err := config.GetAllConnections()
	if err != nil {
		color.Red("Error getting connections: %v", err)
		return &openRootModel{}
	}

	items := make([]table.Row, len(connections))
	for i, conn := range connections {
		items[i] = table.Row{conn.Name, conn.Type.GetMainType(), conn.Type.GetSubtype()}
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
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("212")).
		Bold(true)
	t.SetStyles(s)

	onSelect := func(selected string) tea.Msg {
		conn, err := config.GetConnectionByName(selected)
		if err != nil {
			return fmt.Errorf("error getting connection %s: %w", selected, err)
		}

		return ui.TransitionToShellMsg{
			Connection: conn,
		}
	}

	tblModel := ui.NewTableModel(t, onSelect, false)

	return &openRootModel{
		step:       stepPick,
		tableModel: tblModel,
	}
}

// Init initializes the openRootModel.
func (m *openRootModel) Init() tea.Cmd {
	return m.tableModel.Init()
}

// Update handles messages and updates the model state.
func (m *openRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ui.TransitionToShellMsg:
		m.shellModel = shell.NewShellModel(msg.Connection)
		m.step = stepShell
		return m.shellModel, m.shellModel.Init()

	case error:
		color.Red("Error: %v", msg)
		return m, nil
	}

	if m.step == stepPick {
		var cmd tea.Cmd
		updatedTable, cmd := m.tableModel.Update(msg)
		m.tableModel = updatedTable.(tableModel)
		return m, cmd
	}

	if m.step == stepShell && m.shellModel != nil {
		var cmd tea.Cmd
		m.shellModel, cmd = m.shellModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current view based on the model state.
func (m *openRootModel) View() string {
	if m.step == stepPick {
		return m.tableModel.View()
	}

	if m.step == stepShell && m.shellModel != nil {
		return m.shellModel.View()
	}

	return "No view"
}
