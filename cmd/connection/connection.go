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

	return cmd
}
