package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"

	"github.com/prompt-ops/cli/cmd2/common"
)

func ConnectKubernetes() *cobra.Command {
	var name, context string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "Connect to a Kubernetes cluster",
		Long:  "Connect to a Kubernetes cluster",
		Example: `
# Connect to a Kubernetes cluster
# --name (or -n) is the name of the connection that is being created.
# --context (or -c) is the Kubernetes context to connect to. You can leave this empty if you want to connect to the current context.
pops connect kubernetes --name my-k8s-conn --context my-cluster-context

# Connect to a Kubernetes cluster (interactive)
# --interactive (or -i) flag will run the command in interactive mode.
pops connect kubernetes --interactive
`,
		Run: func(cmd *cobra.Command, args []string) {
			k8sconn, err := NewKubernetesConnection()
			if err != nil {
				fmt.Println(err)
				return
			}

			if interactive {
				k8sconn.handleKubernetesConnectionInteractive()
			} else {
				k8sconn.handleKubernetesConnectionNonInteractive(name, context)
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the Kubernetes connection")
	cmd.Flags().StringVarP(&context, "context", "c", "", "Kubernetes context")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")

	return cmd
}

func (k8sconn *KubernetesConnection) handleKubernetesConnectionInteractive() {
	availableContexts, err := k8sconn.AvailableContexts()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	textInputModel := common.NewTextInputModel("Enter connection name")
	p := tea.NewProgram(textInputModel)
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	connectionName := strings.TrimSpace(m.(common.TextInputModel).Value())
	k8sconn.SetName(connectionName)

	listSelectModel := common.NewListSelectModel("Select Kubernetes context", availableContexts)
	p = tea.NewProgram(listSelectModel)
	m, err = p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	context := strings.TrimSpace(m.(common.ListSelectModel).Selected())
	k8sconn.SetContext(context)

	k8sconn.handleKubernetesConnectionNonInteractive(connectionName, context)
}

func (k8sconn *KubernetesConnection) handleKubernetesConnectionNonInteractive(connectionName, context string) {
	if connectionName == "" {
		fmt.Println("Error: connection-name is required")
		return
	}

	if context == "" {
		currentContext, err := k8sconn.CurrentContext()
		if err != nil {
			fmt.Printf("Error getting current context: %v\n", err)
			return
		}
		context = currentContext
		fmt.Printf("No context provided. Using current context: '%s'\n", context)
	} else {
		fmt.Printf("Connecting to Kubernetes context '%s'\n", context)
	}

	// Save the connection details
	err := common.SaveConnection(common.Connection{
		Name: connectionName,
		Type: "kubernetes",
	})
	if err != nil {
		fmt.Printf("Error saving connection: %v\n", err)
		return
	}

	fmt.Printf("Connection '%s' saved successfully\n", connectionName)

	k8sconn.startShell()
}

func (k8sconn *KubernetesConnection) startShell() error {
	color.Cyan("Starting **pops** interactive shell...")

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	historyFile := filepath.Join(os.TempDir(), ".pops_history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	// Main loop
	for {
		// Get the current context before each prompt
		currentContext, err := k8sconn.CurrentContext()
		if err != nil {
			color.Red("Error getting current context: %s", err)
			return err
		}

		// Prompt the user for input
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

		// Check if the prompt requests a kubectl command or information
		isCommand, err := k8sconn.isCommandRequested(input)
		if err != nil {
			color.Red("Error checking if command is requested: %s", err)
			continue
		}

		if !isCommand {
			color.Yellow("Sorry, I can only process kubectl commands.")
			continue
		}

		kubectlcmd, err := k8sconn.getKubectlCommand(input)
		if err != nil {
			color.Red("Error getting kubectl command: %s", err)
			continue
		}

		if kubectlcmd == "" {
			color.Red("Sorry, I didn't understand that command.")
			continue
		}

		// Warn the user and ask for confirmation
		color.Yellow("The following command will be executed: %s", kubectlcmd)
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
		output, err := exec.Command("sh", "-c", kubectlcmd).CombinedOutput()
		if err != nil {
			color.Red("Error: %s", err)
			color.Red("Command output: %s", string(output))
		} else {
			color.Green("Running kubectl command: %s", kubectlcmd)
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
