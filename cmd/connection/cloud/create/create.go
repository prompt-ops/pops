package create

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/prompt-ops/cli/cmd/connection/cloud/create/models"
)

var (
	name     string
	provider string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud connection",
	Long:  "Create a new cloud connection. This command can be run interactively or non-interactively.",
	Example: `
# Interactive mode
pops connection cloud create

# Non-interactive mode
pops connection cloud create --name my-cloud-connection --provider azure
`,
	Run: func(cmd *cobra.Command, args []string) {
		if name != "" && provider != "" {
			handleCloudConnection(name, provider)
		} else {
			handleCloudConnectionInteractive()
		}
	},
}

func init() {
	CreateCmd.Flags().StringVar(&name, "name", "", "Name of the cloud connection")
	CreateCmd.Flags().StringVar(&provider, "provider", "", "Cloud provider (e.g., azure)")
}

func handleCloudConnection(name, provider string) {
	fmt.Printf("Creating cloud connection %s with provider %s\n", name, provider)
}

func handleCloudConnectionInteractive() {
	flow := models.NewFlowModel()
	p := tea.NewProgram(flow)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
