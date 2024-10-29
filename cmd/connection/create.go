package connection

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create [type] [name]",
		Short:   "Create a new connection",
		Long:    "Create a new connection of the specified type with the given name.",
		Example: `pops connection create kubernetes my-k8s-connection`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			connectionType := args[0]
			connectionName := args[1]

			switch strings.ToLower(connectionType) {
			case "kubernetes":
				handleKubernetesConnection(connectionName)
			case "rdbms":
				handleRDBMSConnection(connectionName)
			default:
				color.Red("Unknown connection type: %s", connectionType)
			}
		},
	}

	return createCmd
}
