package cmd

import (
	"os"

	"github.com/spf13/cobra"

	connection "github.com/prompt-ops/cli/cmd/connection"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pops",
	Short: "your tech assistant (inspired by Lucius Fox)",
	Long:  `PromptOps is your personal tech assistant, inspired by the ingenuity of Lucius Fox.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pops.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// Import the new command
	// rootCmd.AddCommand(newCmd)

	// Add connection commands
	// Create a new connection command
	rootCmd.AddCommand(connection.ConnectionRootCmd)
}
