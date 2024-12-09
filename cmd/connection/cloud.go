package connection

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/peterh/liner"

	config "github.com/prompt-ops/cli/cmd/config"
)

// handleCloudConnection handles the creation of a cloud connection
func handleCloudConnection(name string) {
	reader := bufio.NewReader(os.Stdin)
	color.Cyan("Select cloud provider (currently only Azure is supported): ")
	provider, _ := reader.ReadString('\n')
	provider = strings.TrimSpace(strings.ToLower(provider))

	switch provider {
	case "azure":
		handleAzureConnection(name)
	default:
		color.Red("Unsupported cloud provider: %s", provider)
	}
}

// handleAzureConnection handles the creation of an Azure connection
func handleAzureConnection(name string) {
	// Check if the user is logged in to Azure
	if !isAzureLoggedIn() {
		color.Red("Azure CLI is not logged in. Please run `az login` before creating an Azure connection.")
		return
	}

	// Save the connection details
	conn := config.Connection{
		Type: "azure",
		Name: name,
	}
	if err := config.SaveConnection(conn); err != nil {
		color.Red("Error saving connection: %v", err)
		return
	}

	color.Green("Azure connection '%s' created successfully.", name)

	// Optionally, fetch and display Azure resources
	resourceGroups, err := listAzureResourceGroups()
	if err != nil {
		color.Red("Error fetching Azure resource groups: %v", err)
		return
	}

	if len(resourceGroups) > 0 {
		color.Green("Azure Resource Groups:")
		for _, rg := range resourceGroups {
			color.Yellow("- " + rg)
		}
	} else {
		color.Yellow("No Azure Resource Groups found.")
	}

	color.Cyan("Starting **pops** interactive shell. Type your command, or type 'exit' to quit.")

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	historyFile := filepath.Join(os.TempDir(), ".pops_history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		currentContext := "Azure"
		prompt := fmt.Sprintf("[%s] > ", currentContext)
		input, err := line.Prompt(prompt)
		if err == liner.ErrPromptAborted {
			color.Cyan("Exiting PromptOps shell.")
			break
		} else if err != nil {
			color.Red("Error reading line: %s", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		line.AppendHistory(input)

		if input == "exit" {
			color.Cyan("Exiting PromptOps shell.")
			break
		}

		parsedResponse, err := getCommand(input, CloudCommand, "")
		if err != nil {
			color.Red("Error: %s", err)
			continue
		}

		if parsedResponse.Command == "" {
			color.Red("Sorry, I didn't understand that command.")
			continue
		}

		// Warn the user and ask for confirmation
		color.Yellow("The following command will be executed: %s", parsedResponse.Command)
		confirmationPrompt := "Do you want to proceed? (Y/n): "
		confirmation, err := line.Prompt(confirmationPrompt)
		if err == liner.ErrPromptAborted {
			color.Cyan("Command execution aborted.")
			continue
		} else if err != nil {
			color.Red("Error reading confirmation: %s", err)
			continue
		}

		confirmation = strings.TrimSpace(confirmation)
		if strings.ToLower(confirmation) != "y" {
			color.Cyan("Command execution aborted.")
			continue
		}

		// Execute the kubectl command
		output, err := exec.Command("sh", "-c", parsedResponse.Command).CombinedOutput()
		if err != nil {
			color.Red("Error: %s", err)
			color.Red("Command output: %s", string(output))
		} else {
			color.Green("Running az command: %s", parsedResponse.Command)
			fmt.Println(string(output))
		}
	}

	if f, err := os.Create(historyFile); err != nil {
		color.Red("Error writing history file: %s", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

// isAzureLoggedIn checks if the user is logged in to Azure CLI
func isAzureLoggedIn() bool {
	cmd := exec.Command("az", "account", "show")
	err := cmd.Run()
	return err == nil
}

// listAzureResourceGroups retrieves a list of Azure resource groups
func listAzureResourceGroups() ([]string, error) {
	cmd := exec.Command("az", "group", "list", "--query", "[].name", "-o", "tsv")
	fmt.Println("Running command: ", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	resourceGroups := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(resourceGroups) == 1 && resourceGroups[0] == "" {
		return []string{}, nil
	}

	return resourceGroups, nil
}
