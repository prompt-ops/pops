package azure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/peterh/liner"

	"github.com/prompt-ops/cli/pkg/ai"
	config "github.com/prompt-ops/cli/pkg/config"
)

type AzureConnection struct {
	Name string
}

func NewAzureConnection(name string) *AzureConnection {
	return &AzureConnection{
		Name: name,
	}
}

func (a *AzureConnection) CreateConnection(name string) error {
	// Check if the user is logged in to Azure
	err := a.CheckAuthentication()
	if err != nil {
		color.Red("Azure authentication error: %v", err)
		return err
	}

	// Save the connection details
	conn := config.Connection{
		Type:    a.Type(),
		SubType: a.SubType(),
		Name:    name,
	}
	if err := config.SaveConnection(conn); err != nil {
		color.Red("Error saving connection: %v", err)
		return err
	}

	color.Green("Azure connection '%s' created successfully.", name)

	return nil
}

func (a *AzureConnection) CheckAuthentication() error {
	cmd := exec.Command("az", "account", "show")
	return cmd.Run()
}

func (a *AzureConnection) PrintInitialContext() error {
	// Optionally, fetch and display Azure resources
	resourceGroups, err := listAzureResourceGroups()
	if err != nil {
		color.Red("Error fetching Azure resource groups: %v", err)
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Resource Groups"})
	for _, resourceGroup := range resourceGroups {
		table.Append([]string{
			resourceGroup,
		})
	}
	table.Render()

	return nil
}

func (a *AzureConnection) MainLoop() error {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	historyFile := filepath.Join(os.TempDir(), ".pops_history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		currentContext := a.SubType()
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

		parsedResponse, err := ai.GetCommand(input, a.CommandType(), "")
		if err != nil {
			color.Red("Error: %s", err)
			continue
		}

		if parsedResponse.Command == "" {
			color.Red("Sorry, I didn't understand that prompt.")
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

	return nil
}

func (a *AzureConnection) Type() string {
	return "cloud"
}

func (a *AzureConnection) SubType() string {
	return "azure"
}

func (a *AzureConnection) CommandType() string {
	return "Azure CLI `az` command"
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
