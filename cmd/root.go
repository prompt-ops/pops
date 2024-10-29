package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"

	connection "github.com/prompt-ops/cli/cmd/connection"
	session "github.com/prompt-ops/cli/cmd/session"
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
	// Create a large text with the LetterStyle from the standard theme.
	// This is useful for creating title screens.
	// pterm.DefaultBigText.WithLetters(putils.LettersFromString("Prompt-Ops")).Render()

	// Create a large text with differently colored letters.
	// Here, the first letter 'P' is colored cyan and the rest 'Term' is colored light magenta.
	// This can be used to highlight specific parts of the text.
	// pterm.DefaultBigText.WithLetters(
	// 	putils.LettersFromStringWithStyle("Prompt", pterm.FgCyan.ToStyle()),
	// 	putils.LettersFromStringWithStyle("Ops", pterm.FgLightMagenta.ToStyle()),
	// ).Render()

	// Create a large text with a specific RGB color.
	// This can be used when you need a specific color that is not available in the standard colors.
	// Here, the color is gold (RGB: 255, 215, 0).
	pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithRGB("Prompt-Ops", pterm.NewRGB(255, 215, 0)),
	).Render()

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
	rootCmd.AddCommand(session.SessionRootCmd)
}
