package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/prompt-ops/pops/cmd/pops/app"

	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run cmd/docgen/main.go <output directory>")
	}

	output := os.Args[1]
	_, err := os.Stat(output)
	if os.IsNotExist(err) {
		err = os.Mkdir(output, 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(app.NewRootCommand(), output, frontmatter, link)
	if err != nil {
		log.Fatal(err) //nolint:forbidigo // this is OK inside the main function.
	}
}

const template = `---
type: docs
title: "%s CLI reference"
linkTitle: "%s"
slug: %s
url: %s
description: "Details on the %s Prompt-Ops CLI command"
---
`

func frontmatter(filename string) string {
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	command := strings.Replace(base, "_", " ", -1)
	url := "/reference/cli/" + strings.ToLower(base) + "/"
	return fmt.Sprintf(template, command, command, base, url, command)
}

func link(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "{{< ref " + strings.ToLower(base) + ".md >}}"
}
