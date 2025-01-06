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

- Available Cloud connection types: Azure and AWS.
- Commands: create, delete, open, list, types.
- Examples:
 * 'pops connection cloud create' creates a connection to a cloud provider.
 * 'pops connection cloud open' opens an existing cloud connection.
 * 'pops connection cloud list' lists all cloud connections.
 * 'pops connection cloud delete' deletes a cloud connection.
 * 'pops connection cloud types' lists all available cloud connection types (for now; Azure and AWS).

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
