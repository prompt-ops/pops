package connection

import (
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
	"github.com/spf13/cobra"
)

func NewConnectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage connections for various providers.",
	}

	cmd.AddCommand(cloud.NewCloudCommand())
	cmd.AddCommand(kubernetes.NewKubernetesCommand())

	// `pops connection list` command
	cmd.AddCommand(newListCmd())

	// `pops connection delete *` command
	// `pops connection delete` is an interactive command
	// `pops connection delete my-connection` deletes a connection by name
	// `pops connection delete --all` deletes all connections
	cmd.AddCommand(newDeleteCmd())

	// `pops connection open` command
	cmd.AddCommand(newOpenCmd())

	// `pops connection create` is an interactive command
	cmd.AddCommand(newCreateCmd())

	// `pops connection types` command
	cmd.AddCommand(newTypesCmd())

	return cmd
}
