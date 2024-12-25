// cmd/root.go
package cmd

import (
	"github.com/prompt-ops/pops/cmd/connection"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pops",
		Short: "Pops is a CLI tool for managing cloud and Kubernetes connections.",
		// Optionally, you can add a Run function or leave it empty if the root command only has subcommands
	}

	// Add subcommands
	cmd.AddCommand(connection.NewConnectionCommand())

	return cmd
}
