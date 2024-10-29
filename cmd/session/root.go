package session

import (
	"os"

	"github.com/spf13/cobra"
)

var SessionRootCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  `Manage sessions of various connections.`,
}

func Execute() {
	err := SessionRootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	SessionRootCmd.AddCommand(newCreateCmd())
	SessionRootCmd.AddCommand(newListCmd())
}
