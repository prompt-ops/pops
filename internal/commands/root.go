package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for pops CLI
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pops",
		Short: "PromptOps - AI-powered infrastructure management",
		Long: `PromptOps allows you to interact with your infrastructure using natural language.
Connect to databases, Kubernetes clusters, and cloud resources, then ask questions
or request commands in plain English.`,
	}

	// Add subcommands
	cmd.AddCommand(
		NewConnCmd(),
		NewVersionCmd(),
	)

	return cmd
}