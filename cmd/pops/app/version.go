package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var NewVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Prompt-Ops",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Prompt-Ops version %s\n", version)
	},
}
