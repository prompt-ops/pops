package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesConnection struct {
	name       *string
	context    *string
	clientset  *kubernetes.Clientset
	namespaces []*string
}

func NewKubernetesConnection() (*KubernetesConnection, error) {
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
		context:   nil,
		clientset: clientset,
	}, nil
}

func (k8sconn *KubernetesConnection) SetName(name string) {
	k8sconn.name = &name
}

func (k8sconn *KubernetesConnection) SetContext(context string) {
	k8sconn.context = &context

	// Each time the context is set, we need to fetch the namespaces of the new context.
	k8sconn.fetchNamespaces()
}

func (k8sconn *KubernetesConnection) GetClientset() *kubernetes.Clientset {
	return k8sconn.clientset
}

func (k8sconn *KubernetesConnection) AvailableContexts() ([]string, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Error loading Kubernetes config: %v", err)
	}

	var contexts []string
	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	return contexts, nil
}

func (k8conn *KubernetesConnection) CurrentContext() (string, error) {
	config, err := clientcmd.LoadFromFile(clientcmd.RecommendedHomeFile)
	if err != nil {
		return "", fmt.Errorf("Error loading Kubernetes config: %v", err)
	}

	return config.CurrentContext, nil
}

func (k8sconn *KubernetesConnection) fetchNamespaces() error {
	namespaces, err := k8sconn.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Error fetching namespaces: %v", err)
	}

	var ns []*string
	for _, namespace := range namespaces.Items {
		ns = append(ns, &namespace.Name)
	}
	k8sconn.namespaces = ns

	return nil
}

// This function will ask ChatGPT to tell us if the user requests a command or wants answers to a prompt.
func (k8sconn *KubernetesConnection) isCommandRequested(prompt string) (bool, error) {
	err := godotenv.Load(".env.local")
	if err != nil {
		return false, fmt.Errorf("Error loading .env.local file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return false, fmt.Errorf("OpenAI API key not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Please answer with 'command' if the user prompt wants a command as an answer, or 'information' if the user wants you to provide information."),
			openai.UserMessage(fmt.Sprintf("The user prompt is %s.", string(prompt))),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return false, fmt.Errorf("Error from OpenAI API: %v", err)
	}

	return strings.TrimSpace(chatCompletion.Choices[0].Message.Content) == "command", nil
}

var (
	// defaultSystemMessage is the system message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultSystemMessage = "You are a helpful assistant that translates natural language commands to kubectl command."

	// defaultUserMessage is the user message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultUserMessage = "Translate the following user input to a kubectl command (in your response I only need the kubectl command and nothing else [no quotes or anything like that]): %s"
)

func (k8sconn *KubernetesConnection) getKubectlCommand(prompt string) (string, error) {
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
			openai.SystemMessage(fmt.Sprintf(defaultSystemMessage)),
			openai.UserMessage(fmt.Sprintf(defaultUserMessage, prompt)),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", fmt.Errorf("Error from OpenAI API: %v", err)
	}

	return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
}
