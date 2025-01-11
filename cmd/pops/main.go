package main

import (
	"os"

	"github.com/prompt-ops/pops/cmd/pops/app"
)

func main() {
	err := app.NewRootCommand().Execute()
	if err != nil {
		os.Exit(1) //nolint:forbidigo // this is OK inside the main function.
	}
}
