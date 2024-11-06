package connection

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
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
	conn := Connection{
		Type: "azure",
		Name: name,
	}
	if err := SaveConnection(conn); err != nil {
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
