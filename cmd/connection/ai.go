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
	RDBMSQuery        CommandType = "SQL query"
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
			openai.SystemMessage(fmt.Sprintf("You are a helpful assistant that translates natural language commands to %s.", string(commandType))),
			openai.UserMessage(fmt.Sprintf("Translate the following prompt to a %s but just return the %s (in the response I only need %s and nothing else no quotes or anything like that): %s",
				string(commandType), string(commandType), string(commandType), input)),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		fmt.Printf("Error from OpenAI API: %v", err)
		return "", fmt.Errorf("Error from OpenAI API: %v", err)
	}

	return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
}
