package kubernetes

import (
	"github.com/spf13/cobra"
)

func NewKubernetesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "Manage kubernetes connections.",
	}

	// `pops connection kubernetes create *` commands
	cmd.AddCommand(newCreateCmd())

	// `pops connection kubernetes open *` commands
	cmd.AddCommand(newOpenCmd())

	// `pops connection kubernetes list` command
	cmd.AddCommand(newListCmd())

	// `pops connection kubernetes delete *` commands
	cmd.AddCommand(newDeleteCmd())

	// `pops connection kubernetes types` command
	cmd.AddCommand(newTypesCmd())

	return cmd
}
