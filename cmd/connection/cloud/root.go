package cloud

import (
	"github.com/spf13/cobra"

	"github.com/prompt-ops/cli/cmd/connection/cloud/create"
)

// Entry point for the `pops connection cloud` commands.
// `pops connection cloud create`
// `pops connection cloud list`
// `pops connection cloud types`
// `pops connection cloud delete`

var CloudRootCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Manage cloud connections",
	Long:  "Manage cloud connections, including creating, listing, and deleting connections.",
}

// Add the `createCmd` to the root `cloudCmd`
func init() {
	CloudRootCmd.AddCommand(create.CreateCmd)
}
