package cloud

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/ui"
	cloudui "github.com/prompt-ops/pops/ui/cloud"
	"github.com/spf13/cobra"
)

type createModel struct {
	current tea.Model
}

func initialCreateModel() *createModel {
	return &createModel{
		current: cloudui.NewCreateModel(),
	}
}

// NewCreateModel returns a new createModel
func NewCreateModel() *createModel {
	return initialCreateModel()
}

func (m *createModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.TransitionToShellMsg:
		shell := ui.NewShellModel(msg.Connection)
		return shell, shell.Init()
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m *createModel) View() string {
	return m.current.View()
}

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cloud connection.",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(initialCreateModel())
			if _, err := p.Run(); err != nil {
				panic(err)
			}
		},
	}
	return cmd
}
