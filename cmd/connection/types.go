package connection

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var AvailableConnectionTypes = []string{
	"kubernetes",
	"db",
	"cloud",
}

func GetAvailableConnectionTypes() []string {
	return AvailableConnectionTypes
}

func newListTypesCmd() *cobra.Command {
	listTypesCmd := &cobra.Command{
		Use:   "types",
		Short: "List all available connection types",
		Long:  "List all available connection types that can be created.",
		Run: func(cmd *cobra.Command, args []string) {
			availableTypes := GetAvailableConnectionTypes()

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
