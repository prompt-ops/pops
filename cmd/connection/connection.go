package connection

import (
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/db"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
	"github.com/spf13/cobra"
)

// NewConnectionCommand creates the 'connection' command with descriptions and examples for managing connections.
func NewConnectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage your infrastructure connections using natural language.",
		Long: `
Prompt-Ops manages your infrastructure using natural language.

**Cloud Connection:**
- **Types**: Azure, AWS
- **Commands**: create, delete, open, list, types
- **Example**: 'pops connection cloud create' creates a connection to a cloud provider.

**Database Connection:**
- **Types**: MySQL, PostgreSQL, MongoDB
- **Commands**: create, delete, open, list, types
- **Example**: 'pops connection db create' creates a connection to a database.

**Kubernetes Connection:**
- **Types**: Any available Kubernetes cluster
- **Commands**: create, delete, open, list, types
- **Example**: 'pops connection kubernetes create' creates a connection to a Kubernetes cluster.

More connection types and features are coming soon!`,
		Example: `
- **pops connection create** - Create a connection by selecting from available types.
- **pops connection open** - Open a connection by selecting from available connections.
- **pops connection delete** - Delete a connection by selecting from available connections.
- **pops connection delete --all** - Delete all available connections.
- **pops connection list** - List all available connections.
						`,
	}

	// Add subcommands
	cmd.AddCommand(cloud.NewRootCommand())
	cmd.AddCommand(kubernetes.NewRootCommand())
	cmd.AddCommand(db.NewRootCommand())

	// Add additional commands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newOpenCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newTypesCmd())

	return cmd
}
