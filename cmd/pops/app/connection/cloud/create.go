package cloud

import (
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/connection"
	"github.com/prompt-ops/pops/pkg/ui"
	"github.com/prompt-ops/pops/pkg/ui/cloud"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type createModel struct {
	current tea.Model
}

func initialCreateModel() *createModel {
	return &createModel{
		current: cloud.NewCreateModel(),
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
	var name string
	var provider string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cloud connection.",
		Long: `
Cloud Connection:

- Available Cloud connection types: Azure, AWS, and GCP.
- Commands: create, delete, open, list, types.
- Examples:
 * 'pops connection cloud create' creates a connection interactively.
 * 'pops connection cloud create --name my-azure-conn --provider azure' creates a connection non-interactively.
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Non-interactive mode
			if name != "" && provider != "" {
				err := createCloudConnection(name, provider)
				if err != nil {
					fmt.Printf("Error creating cloud connection: %v\n", err)
					return
				}

				transitionMsg := ui.TransitionToShellMsg{
					Connection: connection.NewCloudConnection(name,
						connection.AvailableCloudConnectionType{
							Subtype: strings.Title(provider),
						},
					),
				}

				p := tea.NewProgram(initialCreateModel())

				// Trying to send the transition message before we start the loop.
				go func() {
					p.Send(transitionMsg)
				}()

				if _, err := p.Run(); err != nil {
					fmt.Printf("Error transitioning to shell: %v\n", err)
				}
			} else {
				// Interactive mode
				p := tea.NewProgram(initialCreateModel())
				if _, err := p.Run(); err != nil {
					fmt.Printf("Error running interactive mode: %v\n", err)
				}
			}
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the cloud connection")
	cmd.Flags().StringVar(&provider, "provider", "", "Cloud provider (azure, aws, gcp)")

	return cmd
}

func createCloudConnection(name, provider string) error {
	name = strings.TrimSpace(name)
	provider = strings.ToLower(strings.TrimSpace(provider))

	if name == "" {
		return fmt.Errorf("connection name cannot be empty")
	}

	var selectedProvider connection.AvailableCloudConnectionType
	for _, p := range connection.AvailableCloudConnectionTypes {
		if strings.ToLower(p.Subtype) == provider {
			selectedProvider = p
			break
		}
	}
	if selectedProvider.Subtype == "" {
		return fmt.Errorf("unsupported cloud provider: %s", provider)
	}

	if config.CheckIfNameExists(name) {
		return fmt.Errorf("connection name '%s' already exists", name)
	}

	connection := connection.NewCloudConnection(name, selectedProvider)
	if err := config.SaveConnection(connection); err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	fmt.Printf("✅ Cloud connection '%s' created successfully with provider '%s'.\n", name, selectedProvider.Subtype)
	return nil
}
