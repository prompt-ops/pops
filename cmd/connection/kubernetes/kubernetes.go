package kubernetes

import (
	"github.com/spf13/cobra"
)

func NewKubernetesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "Manage kubernetes connections.",
	}

	cmd.AddCommand(NewKubernetesCreateCommand())
	cmd.AddCommand(NewKubernetesOpenCommand())

	return cmd
}
