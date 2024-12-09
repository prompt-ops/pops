package connection

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/peterh/liner"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	config "github.com/prompt-ops/cli/cmd/config"
)

type KubernetesConnection struct {
	Context   *string
	Clientset *kubernetes.Clientset
}

func NewKubernetesConnection(selectedContext *string) (*KubernetesConnection, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Error creating Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating Kubernetes client: %v", err)
	}

	return &KubernetesConnection{
		Context:   selectedContext,
		Clientset: clientset,
	}, nil
}

func handleKubernetesConnection(name string) {
	kubeconfig, err := clientcmd.LoadFromFile(clientcmd.RecommendedHomeFile)
	if err != nil {
		color.Red("Error loading kubeconfig file: %v", err)
		return
	}

	contexts := kubeconfig.Contexts
	currentContext := kubeconfig.CurrentContext

	color.Cyan("Select a Kubernetes context to use:")
	color.Cyan("0. None (use current context: %s)", currentContext)
	i := 1
	for contextName := range contexts {
		color.Cyan("%d. %s", i, contextName)
		i++
	}

	reader := bufio.NewReader(os.Stdin)
	color.Cyan("Enter the number of the context to use: ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection) // This will remove both \r and \n

	selectedIndex, err := strconv.Atoi(selection)
	if err != nil || selectedIndex < 0 || selectedIndex >= i {
		color.Red("Invalid selection")
		return
	}

	var selectedContext string
	if selectedIndex == 0 {
		selectedContext = currentContext
	} else {
		i = 1
		for contextName := range contexts {
			if i == selectedIndex {
				selectedContext = contextName
				break
			}
			i++
		}
	}

	// Save the connection details
	conn := config.Connection{
		Type: "kubernetes",
		Name: name,
	}
	if err := config.SaveConnection(conn); err != nil {
		color.Red("Error saving connection: %v", err)
		return
	}

	color.Blue("Creating Kubernetes connection '%s' with context '%s'", name, selectedContext)
	color.Cyan("Starting **pops** interactive shell. Type your command, or type 'exit' to quit.")

	kc, err := NewKubernetesConnection(&selectedContext)

	// Fetch and display the last 10 events from the Kubernetes cluster
	suggestions, err := kc.fetchClusterEvents()
	if err != nil {
		color.Red("Error fetching cluster events: %s", err)
	} else if len(suggestions) > 0 {
		color.Green("Here are some recent events in your cluster:")
		for _, suggestion := range suggestions {
			color.Yellow("- " + suggestion)
		}
	} else {
		color.Yellow("No recent events found in your cluster.")
	}

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	historyFile := filepath.Join(os.TempDir(), ".pops_history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		// Get the current context before each prompt
		currentContext, err := kc.getCurrentContext()
		if err != nil {
			color.Red("Error getting current context: %s", err)
			currentContext = "unknown"
		}

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

		// Example of mapping natural language to a kubectl command using OpenAI
		parsedResponse, err := getCommand(input, KubernetesCommand, "")
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
			color.Green("Running kubectl command: %s", parsedResponse.Command)
			fmt.Println(string(output))
		}

		// // Display suggested next steps
		// if len(parsedResponse.SuggestedSteps) > 0 {
		// 	nextStep, err := selectNextStep(parsedResponse.SuggestedSteps)
		// 	if err != nil {
		// 		color.Red("Error: %s", err)
		// 		continue
		// 	}

		// 	if nextStep != "" {
		// 		color.Green("\nExecuting selected step: %s", nextStep)
		// 		// Reprocess the selected step as a new command
		// 		parsedResponse, err = getCommand(nextStep, KubernetesCommand, "")
		// 		if err != nil {
		// 			color.Red("Error processing selected step: %s", err)
		// 			continue
		// 		}

		// 		output, err = exec.Command("sh", "-c", parsedResponse.Command).CombinedOutput()
		// 		if err != nil {
		// 			color.Red("Error: %s", err)
		// 			color.Red("Command output: %s", string(output))
		// 		} else {
		// 			color.Green("Running kubectl command: %s", parsedResponse.Command)
		// 			fmt.Println(string(output))
		// 		}
		// 	} else {
		// 		color.Yellow("Skipping suggested steps.")
		// 	}
		// } else {
		// 	color.Yellow("No suggested next steps available.")
		// }
	}

	if f, err := os.Create(historyFile); err != nil {
		color.Red("Error writing history file: %s", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

func (kc *KubernetesConnection) fetchClusterEvents() ([]string, error) {
	events, err := kc.Clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		Limit: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("Error fetching events: %v", err)
	}

	var suggestions []string
	for _, event := range events.Items {
		suggestions = append(suggestions, fmt.Sprintf("%s: %s", event.InvolvedObject.Name, event.Message))
	}

	return suggestions, nil
}

func (kc *KubernetesConnection) getCurrentContext() (string, error) {
	config, err := clientcmd.LoadFromFile(clientcmd.RecommendedHomeFile)
	if err != nil {
		return "", fmt.Errorf("Error loading kubeconfig file: %v", err)
	}

	return config.CurrentContext, nil
}

func selectNextStep(suggestions []string) (string, error) {
	if len(suggestions) == 0 {
		return "", fmt.Errorf("No suggestions available")
	}

	color.Cyan("\nPlease select one of the suggested next steps:")
	for i, step := range suggestions {
		color.Yellow("%d. %s", i+1, step)
	}

	reader := bufio.NewReader(os.Stdin)
	color.Cyan("Enter the number of the step you want to execute (or press Enter to skip): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("Error reading input: %v", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil // Skip selection
	}

	selectedIndex, err := strconv.Atoi(input)
	if err != nil || selectedIndex < 1 || selectedIndex > len(suggestions) {
		return "", fmt.Errorf("Invalid selection")
	}

	return suggestions[selectedIndex-1], nil
}
