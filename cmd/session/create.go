package session

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create [type] [name]",
		Short:   "Create a new session for a connection",
		Long:    "Create a new session for a connection with the given name.",
		Example: `pops session create my-new-session --connection my-k8s-connection`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			sessionName := args[0]

			connectionName, _ := cmd.Flags().GetString("connection")
			if connectionName == "" {
				fmt.Println("Connection name is required")
				return
			}

			createSession(sessionName, connectionName)
		},
	}

	createCmd.Flags().StringP("connection", "c", "", "Name of the connection")

	return createCmd
}

func createSession(sessionName, connectionName string) {
	// Logic to create a new session for the given connection
	fmt.Printf("New session '%s' created for connection: %s\n", sessionName, connectionName)
}
