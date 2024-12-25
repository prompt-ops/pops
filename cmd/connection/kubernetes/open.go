package kubernetes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/cli/ui"
	k8sui "github.com/prompt-ops/cli/ui/kubernetes"
	"github.com/spf13/cobra"
)

type openModel struct {
	current tea.Model
}

func initialOpenModel() *openModel {
	return &openModel{
		current: k8sui.NewKubernetesOpenModel(),
	}
}

func (m *openModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m *openModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case ui.TransitionToShellMsg:
		m.current = ui.NewShellModel(msg.(ui.TransitionToShellMsg).Connection)
		return m, m.current.Init()
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m *openModel) View() string {
	return m.current.View()
}

func NewKubernetesOpenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Create a new Kubernetes connection.",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(initialOpenModel())
			if _, err := p.Run(); err != nil {
				panic(err)
			}
		},
	}
	return cmd
}
