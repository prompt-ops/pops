package session

import (
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all sessions that have been set up.",
		Run: func(cmd *cobra.Command, args []string) {
			connectionName, _ := cmd.Flags().GetString("connection")

			var sessions []Session
			var err error

			if connectionName != "" {
				sessions, err = ListSessionsByConnection(connectionName)
				if err != nil {
					color.Red("Error listing sessions: %v", err)
					return
				}
			} else {
				sessions, err = ListSessions()
				if err != nil {
					color.Red("Error listing sessions: %v", err)
					return
				}
			}

			if len(sessions) == 0 {
				color.Cyan("No sessions found.")
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Connection"})

			for _, session := range sessions {
				table.Append([]string{
					session.Name,
					session.Connection.Name,
				})
			}

			table.Render()
		},
	}

	return listCmd
}
