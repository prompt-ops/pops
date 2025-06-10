package main

import (
	"os"

	"github.com/prompt-ops/pops/internal/commands"
)

func main() {
	cmd := commands.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
