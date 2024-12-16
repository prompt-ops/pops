package kubernetes

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type CommandType string

const (
	KubernetesCommand CommandType = "kubectl command"
	RDBMSQuery        CommandType = "PostgreSQL SQL query"
	CloudCommand      CommandType = "Azure `az` command"
)

var (
	// defaultSystemMessage is the system message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultSystemMessage = `You are a helpful assistant that translates natural language commands to %s. 
	This is how your response structure MUST be like:
	
	Command: kubectl get pods
	Suggested next steps:
	1. Describe one of the pods.
	2. View logs of one of the pods.

	Command: SELECT * FROM table_name;
	Suggested next steps:
	1. Filter the results based on a specific condition.
	2. Join this table with another table.

	Command: az vm list
	Suggested next steps:
	1. Start a specific VM.
	2. Stop a specific VM.
	
	Do not include any Markdown-type formatting. Only provide plain text.`

	// defaultUserMessage is the user message that is sent to the OpenAI API to help it understand the context of the user's input.
	defaultUserMessage = "User prompt: %s. Additional context: %s"
)

// ParsedResponse holds the parsed command and suggested next steps.
type ParsedResponse struct {
	Command        string
	SuggestedSteps []string
}

func getCommand(input string, commandType CommandType, extraContext string) (ParsedResponse, error) {
	err := godotenv.Load(".env.local")
	if err != nil {
		return ParsedResponse{}, fmt.Errorf("Error loading .env.local file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return ParsedResponse{}, fmt.Errorf("OpenAI API key not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(fmt.Sprintf(defaultSystemMessage, string(commandType))),
			openai.UserMessage(fmt.Sprintf(defaultUserMessage, input, extraContext)),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return ParsedResponse{}, fmt.Errorf("Error from OpenAI API: %v", err)
	}

	response := strings.TrimSpace(chatCompletion.Choices[0].Message.Content)
	fmt.Printf("AI response: %s\n", response)
	parsedResponse, err := parseResponse(response, commandType)
	if err != nil {
		return ParsedResponse{}, err
	}

	return parsedResponse, nil
}

// parseResponse processes the AI response to extract the command and suggested next steps.
func parseResponse(response string, commandType CommandType) (ParsedResponse, error) {
	parsed := ParsedResponse{}

	// Split the response into lines for parsing
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Command:") {
			parsed.Command = strings.TrimSpace(strings.TrimPrefix(line, "Command:"))
		} else if strings.HasPrefix(line, "Suggested next steps:") {
			// Parse the suggestions
			suggestions := parseSuggestions(lines)
			parsed.SuggestedSteps = suggestions
			break
		}
	}

	// Validate the extracted command
	if commandType == KubernetesCommand && !strings.HasPrefix(parsed.Command, "kubectl") {
		return ParsedResponse{}, fmt.Errorf("Invalid Kubernetes command: %s", parsed.Command)
	}
	if commandType == RDBMSQuery {
		fmt.Println(parsed.Command)
		parsed.Command = cleanSQLQuery(parsed.Command)
		if !isValidSQLQuery(parsed.Command) {
			return ParsedResponse{}, fmt.Errorf("Invalid SQL query: %s", parsed.Command)
		}
	}

	return parsed, nil
}

// parseSuggestions extracts the suggestions from the response.
func parseSuggestions(lines []string) []string {
	var suggestions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Match numbered suggestions (e.g., "1. Describe one of the pods")
		if matched, _ := regexp.MatchString(`^\d+\.\s+`, line); matched {
			suggestions = append(suggestions, line)
		}
	}
	return suggestions
}

// cleanSQLQuery processes the SQL query to remove unnecessary formatting and enforce standards.
func cleanSQLQuery(query string) string {
	query = strings.TrimPrefix(query, "Query: ")
	query = strings.TrimPrefix(query, "```sql")
	query = strings.TrimSuffix(query, "```")
	query = strings.TrimSpace(query)
	return quoteCamelCaseColumns(query)
}

// isValidSQLQuery checks if the query starts with a valid SQL command.
func isValidSQLQuery(query string) bool {
	fmt.Printf("Query: %s\n", query)
	validPrefixes := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(strings.ToUpper(query), prefix) {
			return true
		}
	}
	return false
}

func quoteCamelCaseColumns(query string) string {
	// Regular expression to match camel case column names
	re := regexp.MustCompile(`\b([a-z]+[A-Z][a-zA-Z0-9]*)\b`)
	return re.ReplaceAllStringFunc(query, func(match string) string {
		return fmt.Sprintf(`"%s"`, match)
	})
}
