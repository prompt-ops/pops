package k8s

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubernetes",
		Aliases: []string{"k8s"},
		Short:   "Manage kubernetes connections.",
		Long: `
Kubernetes Connection:

- Available Kubernetes connection types: All Kubernetes clusters defined in your configuration.
- Commands: create, delete, open, list, types.
- Examples:
 * 'pops conn k8s create' creates a connection to a Kubernetes cluster.
 * 'pops conn k8s open' opens an existing Kubernetes connection.
 * 'pops conn k8s list' lists all Kubernetes connections.
 * 'pops conn k8s delete' deletes a Kubernetes connection.
 * 'pops conn k8s types' lists all available Kubernetes connection types (for now; all Kubernetes clusters defined in your configuration).

More connection types and features are coming soon!`,
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
