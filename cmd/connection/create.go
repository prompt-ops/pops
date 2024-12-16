package connection

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/prompt-ops/cli/cmd/connection/db"
	"github.com/prompt-ops/cli/cmd/connection/kubernetes"
)

func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create [type] [name]",
		Short:   "Create a new connection",
		Long:    "Create a new connection of the specified type with the given name.",
		Example: `pops connection create kubernetes my-k8s-connection`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			connectionType := strings.ToLower(args[0])
			connectionName := args[1]

			switch connectionType {
			case "kubernetes":
				kubernetes.HandleKubernetesConnection(connectionName)
			case "db":
				db.HandleDatabaseConnection(connectionName)
			default:
				color.Red("Unknown connection type: %s", connectionType)
			}
		},
	}

	return createCmd
}
