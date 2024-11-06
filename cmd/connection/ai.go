package connection

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type CommandType string

const (
	KubernetesCommand CommandType = "kubectl command"
	RDBMSQuery        CommandType = "PostgreSQL SQL query"
)

var (
	// defaultSystemMessage is the system message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultSystemMessage = "You are a helpful assistant that translates natural language commands to %s."

	// defaultUserMessage is the user message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultUserMessage = "Translate the following user input to a %s (in your response I only need %s and nothing else [no quotes or anything like that]): %s"
)

func getCommand(input string, commandType CommandType) (string, error) {
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
			openai.SystemMessage(fmt.Sprintf(defaultSystemMessage, string(commandType))),
			openai.UserMessage(fmt.Sprintf(defaultUserMessage, string(commandType), string(commandType), input)),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", fmt.Errorf("Error from OpenAI API: %v", err)
	}

	return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
}
