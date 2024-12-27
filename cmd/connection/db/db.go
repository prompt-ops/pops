package db

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage db connections.",
	}

	// `pops connection db create *` commands
	cmd.AddCommand(newCreateCmd())

	// `pops connection db open *` commands
	cmd.AddCommand(newOpenCmd())

	// `pops connection db list` command
	cmd.AddCommand(newListCmd())

	// `pops connection db delete *` commands
	cmd.AddCommand(newDeleteCmd())

	// `pops connection db types` command
	cmd.AddCommand(newTypesCmd())

	return cmd
}
