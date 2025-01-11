package connection

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

// newOpenCmd creates the open command
func newOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open",
		Short: "Open a connection",
		Long:  "Open a connection",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				conn, err := config.GetConnectionByName(args[0])
				if err != nil {
					color.Red("Error getting connection: %v", err)
					return
				}

				shell := ui.NewShellModel(conn)
				p := tea.NewProgram(shell)
				if _, err := p.Run(); err != nil {
					color.Red("Error opening shell UI: %v", err)
				}
			} else {
				root := NewOpenRootModel()
				p := tea.NewProgram(root)
				if _, err := p.Run(); err != nil {
					color.Red("Error: %v", err)
					return
				}
			}
		},
	}

	return openCmd
}

// The root model for "open"
type openRootModel struct {
	step       openStep
	tableModel tableModel
	shellModel tea.Model
}

// openStep enumerates which sub-UI is active.
type openStep int

const (
	stepPick openStep = iota
	stepShell
)

// connectionSelectedMsg signals that the user picked a connection
type connectionSelectedMsg struct {
	Conn connection.Connection
}

type tableModel interface {
	tea.Model
	Selected() string
}

// NewOpenRootModel creates the root model in "pick" mode
func NewOpenRootModel() *openRootModel {
	// Build the table
	connections, err := config.GetAllConnections()
	if err != nil {
		// In a real app, handle this more gracefully
		color.Red("Error getting connections: %v", err)
		// Return an empty model or handle error
		return &openRootModel{}
	}

	// Prepare rows
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

	// Style the table
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

	onSelect := func(selected string) tea.Msg {
		conn, err := config.GetConnectionByName(selected)
		if err != nil {
			return func() tea.Msg {
				return fmt.Errorf("error getting connection %s: %w", selected, err)
			}
		}

		return func() tea.Msg {
			return connectionSelectedMsg{
				Conn: conn,
			}
		}
	}

	tblModel := ui.NewTableModel(t, onSelect)

	return &openRootModel{
		step:       stepPick,
		tableModel: tblModel,
	}
}

// Init initializes whichever sub-UI we’re starting with
func (m *openRootModel) Init() tea.Cmd {
	// Start with the table model
	return m.tableModel.Init()
}

// Update handles messages.
// If we get a `connectionSelectedMsg`, we switch to the shell.
func (m *openRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case connectionSelectedMsg:
		fmt.Println("Selected connection:", msg.Conn.Name)

		// The user picked a connection from the table. Build the shell model:
		m.shellModel = ui.NewShellModel(msg.Conn)
		m.step = stepShell
		// Return the shell model as the new top-level model with its Init.
		return m.shellModel, m.shellModel.Init()

	// If the table tried to look up the connection and failed
	case error:
		color.Red("Error: %v", msg)
		// You could handle it, or just stay in the table.

	}

	// If we’re not transitioning, update the current sub-UI
	if m.step == stepPick {
		var cmd tea.Cmd
		updatedTable, cmd := m.tableModel.Update(msg)
		// Because tableModel is an interface, we need to store it back:
		m.tableModel = updatedTable.(tableModel)
		return m, cmd
	}

	// If we’re in shell mode, pass the message to the shell
	if m.step == stepShell && m.shellModel != nil {
		var cmd tea.Cmd
		m.shellModel, cmd = m.shellModel.Update(msg)
		return m, cmd
	}

	// Default fallback
	return m, nil
}

// View renders whichever sub-UI we’re using
func (m *openRootModel) View() string {
	if m.step == stepPick {
		return m.tableModel.View()
	}
	if m.step == stepShell && m.shellModel != nil {
		return m.shellModel.View()
	}
	return "No view"
}
