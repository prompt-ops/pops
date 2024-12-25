package cloud

import (
	"github.com/spf13/cobra"
)

func NewCloudCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud connections.",
	}

	// Add subcommands
	cmd.AddCommand(NewCloudCreateCommand())
	cmd.AddCommand(NewCloudOpenCommand())

	return cmd
}
