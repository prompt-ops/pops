package cloud

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud connections.",
	}

	// `pops connection cloud create *` commands
	cmd.AddCommand(newCreateCmd())

	// `pops connection cloud open *` commands
	cmd.AddCommand(newOpenCmd())

	// `pops connection cloud list` command
	cmd.AddCommand(newListCmd())

	// `pops connection cloud delete *` commands
	cmd.AddCommand(newDeleteCmd())

	// `pops connection cloud types` command
	cmd.AddCommand(newTypesCmd())

	return cmd
}
