package cloud

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/cli/ui"
	cloudui "github.com/prompt-ops/cli/ui/cloud"
	"github.com/spf13/cobra"
)

type createModel struct {
	current tea.Model
}

func initialCreateModel() *createModel {
	return &createModel{
		current: cloudui.NewCloudCreateModel(),
	}
}

func (m *createModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case ui.TransitionToShellMsg:
		shellModel := ui.NewShellModel(msg.(ui.TransitionToShellMsg).Connection)
		m.current = shellModel
		return m, shellModel.Init()
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m *createModel) View() string {
	return m.current.View()
}

func NewCloudCreateCommand() *cobra.Command {
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
