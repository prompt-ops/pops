package connection

import (
	"github.com/fatih/color"
	config "github.com/prompt-ops/pops/config"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a connection or all connections",
		Long:  "Delete a connection or all connections",
		Run: func(cmd *cobra.Command, args []string) {
			// The argument can be the name of the connection or --all:
			// pops connection delete my-connection
			// pops connection delete --all
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if !all && len(args) == 0 {
				color.Red("Please provide the name of the connection to delete or use --all to delete all connections")
				return
			}

			connectionName := ""
			if !all && len(args) == 1 {
				connectionName = args[0]
			}

			if all {
				color.Yellow("Deleting all connections")
				if err := config.DeleteAllConnections(); err != nil {
					color.Red("Error deleting all connections: %v", err)
				}
				color.Blue("Deleted all connections")
				return
			} else {
				color.Yellow("Deleting connection %s", connectionName)
				if err := config.DeleteConnectionByName(connectionName); err != nil {
					color.Red("Error deleting connection: %v", err)
				}
				color.Blue("Deleted connection %s", connectionName)
				return
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all connections")

	return deleteCmd
}
