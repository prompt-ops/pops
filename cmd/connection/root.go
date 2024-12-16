package connection

import (
	"os"

	"github.com/prompt-ops/cli/cmd/connection/cloud"
	"github.com/spf13/cobra"
)

var ConnectionRootCmd = &cobra.Command{
	Use:   "connection",
	Short: "Manage connections",
	Long:  `Manage connections to various services.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := ConnectionRootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add connection commands
	// Create a new connection command
	ConnectionRootCmd.AddCommand(newCreateCmd())
	// List all connections command
	ConnectionRootCmd.AddCommand(newListCmd())
	// Delete a connection or all connections command
	ConnectionRootCmd.AddCommand(newDeleteCmd())
	// List all available connection types command
	ConnectionRootCmd.AddCommand(newListTypesCmd())
	// Add Cloud commands
	ConnectionRootCmd.AddCommand(cloud.CloudRootCmd)
}
