package db

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage database connections.",
		Long: `
Database Connection:

- Available Database connection types: MySQL, PostgreSQL, and MongoDB.
- Commands: create, delete, open, list, types.
- Examples:
 * 'pops connection db create' creates a connection to a database.
 * 'pops connection db open' opens an existing database connection.
 * 'pops connection db list' lists all database connections.
 * 'pops connection db delete' deletes a database connection.
 * 'pops connection db types' lists all available database connection types (for now; MySQL, PostgreSQL, and MongoDB).

More connection types and features are coming soon!`,
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
