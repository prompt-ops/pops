// cmd/root.go
package cmd

import (
	"github.com/prompt-ops/pops/cmd/connection"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pops",
		Short: "Prompt-Ops manages your infrastructure using natural language.",
	}

	// Add subcommands
	cmd.AddCommand(connection.NewConnectionCommand())

	return cmd
}
