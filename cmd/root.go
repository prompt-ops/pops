package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"

	connection "github.com/prompt-ops/cli/cmd/connection"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pops",
	Short: "your tech assistant (inspired by Lucius Fox)",
	Long:  `Prompt-Ops, aka 'pops', lets you talk to your Kubernetes clusters, databases, and cloud providers in plain English.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithRGB("Prompt-Ops", pterm.NewRGB(255, 215, 0)),
	).Render()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(connection.ConnectionRootCmd)
}
