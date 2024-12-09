package session

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	config "github.com/prompt-ops/cli/cmd/config"
)

func newDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a session or all sessions",
		Long:  "Delete a session or all sessions",
		Run: func(cmd *cobra.Command, args []string) {
			// The argument can be the name of the session or --all:
			// pops session delete my-k8s-session
			// pops session delete --all
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				color.Red("Error parsing flags: %v", err)
				return
			}

			if !all && len(args) == 0 {
				color.Red("Please provide the name of the session to delete or use --all to delete all sessions")
				return
			}

			sessionName := ""
			if !all && len(args) == 1 {
				sessionName = args[0]
			}

			if all {
				color.Yellow("Deleting all sessions")
				if err := config.DeleteAllSessions(); err != nil {
					color.Red("Error deleting all sessions: %v", err)
				}
				color.Blue("Deleted all sessions")
				return
			} else {
				color.Yellow("Deleting session %s", sessionName)
				if err := config.DeleteSessionByName(sessionName); err != nil {
					color.Red("Error deleting session: %v", err)
				}
				color.Blue("Deleted session %s", sessionName)
				return
			}
		},
	}

	deleteCmd.Flags().Bool("all", false, "Delete all sessions")

	return deleteCmd
}
