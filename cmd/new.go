package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PromptOps struct {
	clientset *kubernetes.Clientset
}

func NewPromptOps() (*PromptOps, error) {
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
	return &PromptOps{
		clientset: clientset,
	}, nil
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Starts a new interactive shell to interact with the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("Starting **pops** interactive shell. Type your command, or type 'exit' to quit.")

		pops, err := NewPromptOps()
		if err != nil {
			color.Red("Error creating PromptOps: %s", err)
			return
		}

		// Fetch and display the last 10 events from the Kubernetes cluster
		suggestions, err := pops.fetchClusterEvents()
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
			currentContext, err := pops.getCurrentContext()
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
			kubectlCmd, err := mapToKubectl(input)
			if err != nil {
				color.Red("Error: %s", err)
				continue
			}

			if kubectlCmd == "" {
				color.Red("Sorry, I didn't understand that command.")
				continue
			}

			// Warn the user and ask for confirmation
			color.Yellow("The following command will be executed: %s", kubectlCmd)
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
			output, err := exec.Command("sh", "-c", kubectlCmd).CombinedOutput()
			if err != nil {
				color.Red("Error: %s", err)
				color.Red("Command output: %s", string(output))
			} else {
				color.Green("Running kubectl command: %s", kubectlCmd)
				fmt.Println(string(output))
			}
		}

		if f, err := os.Create(historyFile); err != nil {
			color.Red("Error writing history file: %s", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

// mapToKubectl function to map basic natural language to kubectl commands using OpenAI
func mapToKubectl(input string) (string, error) {
	// Load environment variables from .env.local
	err := godotenv.Load(".env.local")
	if err != nil {
		return "", fmt.Errorf("Error loading .env.local file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that translates natural language commands to kubectl commands."),
			openai.UserMessage(fmt.Sprintf("Translate the following natural language command to a kubectl command but just return the kubectl command without the sh part please (just the command starting with kubectl): %s", input)),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		fmt.Printf("Error from OpenAI API: %v", err)
		return "", fmt.Errorf("Error from OpenAI API: %v", err)
	}

	return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
}

// fetchClusterEvents fetches the last 10 events from the Kubernetes cluster
func (pops *PromptOps) fetchClusterEvents() ([]string, error) {
	events, err := pops.clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
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

// getCurrentContext retrieves the current Kubernetes context from the kubeconfig file
func (pops *PromptOps) getCurrentContext() (string, error) {
	config, err := clientcmd.LoadFromFile(clientcmd.RecommendedHomeFile)
	if err != nil {
		return "", fmt.Errorf("Error loading kubeconfig file: %v", err)
	}

	return config.CurrentContext, nil
}
