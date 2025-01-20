package k8s

import (
	"github.com/prompt-ops/pops/pkg/ui"
	k8sui "github.com/prompt-ops/pops/pkg/ui/conn/k8s"
	"github.com/prompt-ops/pops/pkg/ui/shell"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type openModel struct {
	current tea.Model
}

func initialOpenModel() *openModel {
	return &openModel{
		current: k8sui.NewOpenModel(),
	}
}

// NewOpenModel returns a new openModel
func NewOpenModel() *openModel {
	return initialOpenModel()
}

func (m *openModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m *openModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.TransitionToShellMsg:
		shell := shell.NewShellModel(msg.Connection)
		return shell, shell.Init()
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m *openModel) View() string {
	return m.current.View()
}

func newOpenCmd() *cobra.Command {
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
