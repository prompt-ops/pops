package ai

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	// defaultSystemMessage is the system message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultSystemMessage = `You are a helpful assistant that translates natural language commands to %s.
	This is how your response structure MUST be like:

	Command: az vm list
	Suggested next steps:
	1. Start a specific VM.
	2. Stop a specific VM.

	Command: kubectl get pods
	Suggested next steps:
	1. Describe one of the pods.
	2. Delete a specific pod.

	Command: aws ec2 describe-instances
	Suggested next steps:
	1. Start a specific instance.
	2. Stop a specific instance.

	Command: SELECT * FROM table_name;
	Suggested next steps:
	1. Filter the results based on a specific condition.
	2. Join this table with another table.

	Do not include any Markdown-type formatting. Only provide plain text.`
)

// OpenAIModel is the OpenAI implementation of the AIModel interface.
type OpenAIModel struct {
	// apiKey is the API key for the OpenAI API.
	apiKey string
	// client is the OpenAI client.
	client *openai.Client

	// chatModel is the OpenAI Chat Model to use.
	chatModel openai.ChatModel
	// commandType is the type of command to generate.
	commandType string
	// context is the context to provide to the AI.
	context string
}

func NewOpenAIModel(commandType, context string) (*OpenAIModel, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIModel{
		apiKey: apiKey,
		client: client,

		// chatModel is given a default value, but can be changed by the user.
		chatModel: openai.ChatModelGPT4o,

		// commandType and context are set in the constructor by the underlying connection.
		commandType: commandType,
		context:     context,
	}, nil
}

func (o *OpenAIModel) GetName() string {
	return "OpenAI"
}

func (o *OpenAIModel) GetAPIKey() string {
	return o.apiKey
}

func (o *OpenAIModel) SetChatModel(chatModel openai.ChatModel) {
	o.chatModel = chatModel
}

func (o *OpenAIModel) GetChatModel() openai.ChatModel {
	return o.chatModel
}

func (o *OpenAIModel) SetCommandType(commandType string) {
	o.commandType = commandType
}

func (o *OpenAIModel) GetCommandType() string {
	return o.commandType
}

func (o *OpenAIModel) SetContext(context string) {
	o.context = context
}

func (o *OpenAIModel) GetContext() string {
	return o.context
}

func (o *OpenAIModel) GetCommand(prompt string) (*AIResponse, error) {
	chatCompletion, err := o.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			// SystemMessage is used to set the context for the AI.
			openai.SystemMessage(fmt.Sprintf(defaultSystemMessage, o.GetCommandType())),
			// Context is used to provide additional context for the AI.
			openai.SystemMessage(o.GetContext()),
			// UserMessage is the prompt from the user.
			openai.UserMessage(prompt),
		}),
		Model: openai.F(o.GetChatModel()),
	})
	if err != nil {
		return &AIResponse{}, fmt.Errorf("Error from OpenAI API: %v", err)
	}

	response := strings.TrimSpace(chatCompletion.Choices[0].Message.Content)
	fmt.Println("AI Response:", response)

	parsedAIResponse, err := parseResponse(response)
	if err != nil {
		return &AIResponse{}, err
	}

	return &parsedAIResponse, nil
}

// parseResponse processes the AI response to extract the command and suggested next steps.
func parseResponse(response string) (AIResponse, error) {
	parsed := AIResponse{}

	// Split the response into lines for parsing
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Command:") {
			parsed.Command = strings.TrimSpace(strings.TrimPrefix(line, "Command:"))
		} else if strings.HasPrefix(line, "Suggested next steps:") {
			// Parse the suggestions
			suggestions := parseSuggestions(lines)
			parsed.NextSteps = suggestions
			break
		}
	}

	return parsed, nil
}

// parseSuggestions extracts the suggestions from the response.
func parseSuggestions(lines []string) []string {
	var suggestions []string
	re := regexp.MustCompile(`^\d+\.\s+`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Match numbered suggestions (e.g., "1. Describe one of the pods")
		if matched := re.MatchString(line); matched {
			suggestions = append(suggestions, line)
		}
	}
	return suggestions
}
