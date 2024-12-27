package kubernetes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/ui"
	k8sui "github.com/prompt-ops/pops/ui/kubernetes"
	"github.com/spf13/cobra"
)

type createModel struct {
	current tea.Model
}

func initialCreateModel() *createModel {
	return &createModel{
		current: k8sui.NewCreateModel(),
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
	switch msg.(type) {
	case ui.TransitionToShellMsg:
		shellModel := ui.NewShellModel(msg.(ui.TransitionToShellMsg).Connection)
		return shellModel, shellModel.Init()
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
		Short: "Create a new Kubernetes connection.",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(initialCreateModel())
			if _, err := p.Run(); err != nil {
				panic(err)
			}
		},
	}
	return cmd
}
