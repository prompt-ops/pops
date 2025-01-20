package cloud

import (
	"fmt"

	"github.com/prompt-ops/pops/pkg/config"
	"github.com/prompt-ops/pops/pkg/conn"
	"github.com/prompt-ops/pops/pkg/ui"
	"github.com/prompt-ops/pops/pkg/ui/conn/cloud"
	"github.com/prompt-ops/pops/pkg/ui/shell"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type openModel struct {
	current tea.Model
}

func initialOpenModel() *openModel {
	return &openModel{
		current: cloud.NewOpenModel(),
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
	var name string

	cmd := &cobra.Command{
		Use:   "open [connection-name]",
		Short: "Open an existing cloud conn.",
		Long: `Open a cloud connection to access its shell.

You can specify the connection name either as a positional argument or using the --name flag.

Examples:
  pops connection cloud open my-azure-conn
  pops connection cloud open --name my-azure-conn
`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var connectionName string

			// Determine the connection name based on flag and arguments
			if name != "" && len(args) > 0 {
				// If both flag and argument are provided, prioritize the flag
				fmt.Println("Warning: --name flag is provided; ignoring positional argument.")
				connectionName = name
			} else if name != "" {
				// If only flag is provided
				connectionName = name
			} else if len(args) == 1 {
				// If only positional argument is provided
				connectionName = args[0]
			} else {
				// Interactive mode if neither flag nor argument is provided
				p := tea.NewProgram(initialOpenModel())
				if _, err := p.Run(); err != nil {
					fmt.Printf("Error running interactive mode: %v\n", err)
				}
				return
			}

			// Non-interactive mode: Open the specified connection
			conn, err := getConnectionByName(connectionName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			transitionMsg := ui.TransitionToShellMsg{
				Connection: conn,
			}

			p := tea.NewProgram(initialOpenModel())

			// Send the transition message before running the program
			go func() {
				p.Send(transitionMsg)
			}()

			if _, err := p.Run(); err != nil {
				fmt.Printf("Error transitioning to shell: %v\n", err)
			}
		},
	}

	// Define the --name flag
	cmd.Flags().StringVar(&name, "name", "", "Name of the cloud connection")

	return cmd
}

// getConnectionByName retrieves a cloud connection by its name.
// Returns an error if the connection does not exist.
func getConnectionByName(name string) (conn.Connection, error) {
	cloudConnections, err := config.GetConnectionsByType(conn.ConnectionTypeCloud)
	if err != nil {
		return conn.Connection{}, fmt.Errorf("failed to retrieve connections: %w", err)
	}

	for _, conn := range cloudConnections {
		if conn.Name == name {
			return conn, nil
		}
	}

	return conn.Connection{}, fmt.Errorf("connection '%s' does not exist", name)
}
