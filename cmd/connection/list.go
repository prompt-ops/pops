package connection

import (
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	config "github.com/prompt-ops/cli/cmd/config"
)

func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all connections",
		Long:  "List all connections that have been set up.",
		Run: func(cmd *cobra.Command, args []string) {
			connections, err := config.ListConnections()
			if err != nil {
				color.Red("Error listing connections: %v", err)
				return
			}

			if len(connections) == 0 {
				color.Cyan("No connections found.")
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Type", "Sessions"})

			for _, conn := range connections {
				table.Append([]string{
					conn.Name,
					conn.Type,
					strconv.Itoa(len(conn.Sessions)),
				})
			}

			table.Render()
		},
	}

	return listCmd
}
