package app

import (
	"fmt"
	"os"

	"github.com/prompt-ops/pops/cmd/pops/app/connection"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pops",
		Short: "Prompt-Ops manages your infrastructure using natural language.",
	}

	// `pops version` command
	cmd.AddCommand(NewVersionCmd)

	// `pops connection` commands
	cmd.AddCommand(connection.NewConnectionCommand())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "Error: %s\n", err)
		os.Exit(1)
	}
}
