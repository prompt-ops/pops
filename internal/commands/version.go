package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These will be set at build time via ldflags
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// NewVersionCmd creates the version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pops version %s\n", Version)
			fmt.Printf("commit: %s\n", Commit)
			fmt.Printf("built: %s\n", Date)
		},
	}
}