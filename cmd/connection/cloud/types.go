package cloud

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/prompt-ops/cli/pkg/common"
)

func newListTypesCmd() *cobra.Command {
	listTypesCmd := &cobra.Command{
		Use:   "types",
		Short: "List all available cloud connection types",
		Long:  "List all available cloudconnection types that can be created.",
		Run: func(cmd *cobra.Command, args []string) {
			availableTypes := common.GetAvailableCloudConnectionTypes()

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Available Types"})

			for _, availableType := range availableTypes {
				table.Append([]string{
					availableType,
				})
			}

			table.Render()
		},
	}

	return listTypesCmd
}
