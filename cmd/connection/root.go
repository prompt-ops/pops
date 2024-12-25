package connection

import (
	"os"

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
	// ConnectionRootCmd.AddCommand(cloud.CloudRootCmd)
}
