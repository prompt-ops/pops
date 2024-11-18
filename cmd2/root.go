package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"

	connection "github.com/prompt-ops/cli/cmd2/connection"
)

var rootCmd = &cobra.Command{
	Use:   "pops",
	Short: "Prompt-Ops",
	Long:  `Prompt-Ops lets you talk to your things...`,
	Example: `
# Connect to a Kubernetes cluster
pops connect kubernetes --connection-name my-k8s-conn --context my-cluster-context

# Connect to a Kubernetes cluster (interactive)
pops connect kubernetes --interactive

# Connect to a database
pops connect database --connection-name my-db-conn \
	--host my-db-host \
	--port 5432 \
	--username my-db-user \
	--password my-db-pass \
	--database my-db-name \
	--db-type postgres

# Connect to a database (interactive)
pops connect database --interactive

# Connect to a cloud provider
pops connect cloud --connection-name my-cloud-conn \
	--provider azure

# Connect to a cloud provider (interactive)
pops connect cloud --interactive

# List all connections
pops list connections

# Delete a connection
pops delete connection --connection-name my-db-conn

# Start a new session
# A session keeps track of all the interactions with a connection.
pops start session --session-name my-session \
	--connection-name my-db-conn

# Start a new session (interactive)
pops start session --interactive

# Resume an existing session
pops resume session --session-name my-session

# Resume an existing session (interactive)
pops resume session --interactive

# List all sessions
pops list sessions

# Delete a session
pops delete session --session-name my-session
	`,
}

func Execute() {
	pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithRGB("Prompt-Ops", pterm.NewRGB(255, 215, 0)),
	).Render()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add `pops connect` command
	rootCmd.AddCommand(connection.ConnectRootCmd)
}
