package connection

import (
	"os"

	"github.com/spf13/cobra"

	k8sconn "github.com/prompt-ops/cli/cmd2/connection/kubernetes"
)

var ConnectRootCmd = &cobra.Command{
	Use:   "connect",
	Short: "Manage connections",
	Long:  `Manage connections to various services.`,
	Example: `
# Connect to a Kubernetes cluster
pops connect kubernetes --connection-name my-k8s-conn --context my-cluster-context

# Connect to a Kubernetes cluster (interactive)
pops connect kubernetes --interactive

# Connect to a database
pops connect database --connection-name my-db-conn \
	--host my-db-host \
	--port 5432 \
	--username my-db-user \
	--password my-db-pass \
	--database my-db-name \
	--db-type postgres

# Connect to a database (interactive)
pops connect database --interactive

# Connect to a cloud provider
pops connect cloud --connection-name my-cloud-conn \
	--provider azure

# Connect to a cloud provider (interactive)
pops connect cloud --interactive

# List all connections
pops list connections

# Delete a connection
pops delete connection --connection-name my-db-conn
	`,
}

func Execute() {
	err := ConnectRootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// `pops connect kubernetes` command
	kubernetes := k8sconn.ConnectKubernetes()

	ConnectRootCmd.AddCommand(kubernetes)
}
