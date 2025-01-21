package cloud

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud provider connections.",
		Long: `
Cloud Connection:

- Available Cloud connection types: Azure.
- Commands: create, delete, open, list, types.
- Examples:
 * 'pops conn cloud create' creates a connection to a cloud provider.
 * 'pops conn cloud open' opens an existing cloud connection.
 * 'pops conn cloud list' lists all cloud connections.
 * 'pops conn cloud delete' deletes a cloud connection.
 * 'pops conn cloud types' lists all available cloud connection types (for now; Azure).

More connection types and features are coming soon!`,
	}

	// `pops connection cloud create *` commands
	cmd.AddCommand(newCreateCmd())

	// `pops connection cloud open *` commands
	cmd.AddCommand(newOpenCmd())

	// `pops connection cloud list` command
	cmd.AddCommand(newListCmd())

	// `pops connection cloud delete *` commands
	cmd.AddCommand(newDeleteCmd())

	// `pops connection cloud types` command
	cmd.AddCommand(newTypesCmd())

	return cmd
}
