package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

var (
	// defaultSystemMessage is the system message that is sent to the OpenAI API to help it
	// understand the context of the user's input.
	defaultSystemMessage = `
You are a helpful assistant that translates natural language commands to %s.
You must output your response without any code fences or triple backticks.
If you have SQL code or other commands, simply put them after the word “Command:” with no markdown.
For example:

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

No triple backticks. No code fences. Plain text only.
Do not include any Markdown formatting (like triple backticks or bullet points other than the step numbering).
Do not include any additional explanation or text besides what is requested.
No triple backticks. No code fences. Plain text only.
`
)

// OpenAIModel is the OpenAI implementation of the AIModel interface.
type OpenAIModel struct {
	apiKey      string
	client      *openai.Client
	chatModel   openai.ChatModel
	commandType string
	context     string
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
		apiKey:      apiKey,
		client:      client,
		chatModel:   openai.ChatModelGPT4o, // example model
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

// GetCommand calls the OpenAI API with tool calling (if supported), then falls back to text parsing if no tool call is made.
func (o *OpenAIModel) GetCommand(prompt string) (*AIResponse, error) {

	// 1) Define the tool(s) the model may call.
	tools := []openai.ChatCompletionToolParam{
		{
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.F("generateCommand"),
				Description: openai.F("Generate a command and suggested next steps."),
				Parameters: openai.F(shared.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "The command to run, e.g. 'az vm list'",
						},
						"suggestedNextSteps": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "A list of suggested next steps such as '1. Start a specific VM.' etc.",
						},
					},
					"required":             []string{"command", "suggestedNextSteps"},
					"additionalProperties": false,
				}),
				Strict: openai.F(true),
			}),
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
		},
	}

	// 2) Create the chat completion request with tool definitions. Depending on your client, you might use a different way to specify tool calling.
	chatCompletion, err := o.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(fmt.Sprintf(defaultSystemMessage, o.GetCommandType())),
			openai.SystemMessage(o.GetContext()),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(o.GetChatModel()),
		ToolChoice: openai.F[openai.ChatCompletionToolChoiceOptionUnionParam](
			openai.ChatCompletionToolChoiceOptionBehavior(openai.ChatCompletionToolChoiceOptionBehaviorRequired)),
		Tools:       openai.F(tools),
		Temperature: openai.F(0.2),
	})
	if err != nil {
		return nil, fmt.Errorf("error from OpenAI API: %v", err)
	}

	// 3) Check the response. The model might return a direct text answer or a tool call.
	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := chatCompletion.Choices[0]

	fmt.Println("choice: ", choice)

	// If the model returned a tool call, parse the JSON in toolCall.Arguments.
	if choice.Message.ToolCalls != nil {
		return parseToolCalls(choice.Message.ToolCalls)
	}

	// Otherwise, fallback to your existing text-based parsing.
	response := strings.TrimSpace(choice.Message.Content)
	responseStr := stripMarkdownFences(response)
	parsedAIResponse, err := parseResponse(responseStr)
	if err != nil {
		return nil, err
	}

	// Debug/Log
	fmt.Println("command: ", parsedAIResponse.Command)
	fmt.Println("suggested next steps: ", parsedAIResponse.NextSteps)
	return &parsedAIResponse, nil
}

// parseToolCall unmarshals the model's tool call arguments into AIResponse.
func parseToolCalls(toolCalls []openai.ChatCompletionMessageToolCall) (*AIResponse, error) {
	if len(toolCalls) == 0 {
		return nil, fmt.Errorf("no tool calls found")
	}

	for _, toolCall := range toolCalls {
		if toolCall.Function.Name == "generateCommand" {
			var args struct {
				Command            string   `json:"command"`
				SuggestedNextSteps []string `json:"suggestedNextSteps"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call args: %v", err)
			}

			fmt.Println("command from parseToolCalls: ", args.Command)
			fmt.Println("suggested next steps from parseToolCalls: ", args.SuggestedNextSteps)

			return &AIResponse{
				Command:   args.Command,
				NextSteps: args.SuggestedNextSteps,
			}, nil
		}
	}

	return nil, fmt.Errorf("needed tool call could not be found in the tool calls")
}

// parseResponse processes the AI response to extract the command and suggested next steps
// when the model doesn't perform a tool calling.
func parseResponse(response string) (AIResponse, error) {
	parsed := AIResponse{}

	// Split the response into lines for parsing
	lines := strings.Split(response, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "Command:") {
			parsed.Command = strings.TrimSpace(strings.TrimPrefix(line, "Command:"))
		} else if strings.HasPrefix(line, "Suggested next steps:") {
			parsed.NextSteps = parseSuggestions(lines[i+1:])
			break
		}
	}

	return parsed, nil
}

// parseSuggestions extracts the suggestions from the response lines after "Suggested next steps:".
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

func stripMarkdownFences(s string) string {
	// A simple approach: remove any triple backticks
	// This regex will remove ``` and also handle possible variations like ```sql
	re := regexp.MustCompile("```(sql|bash|[a-zA-Z0-9]*)?")
	s = re.ReplaceAllString(s, "")
	return strings.ReplaceAll(s, "```", "")
}
