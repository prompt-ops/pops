package ai

// AIModel defines the interface for different AI providers.
type AIModel interface {
	// GetName returns the name of the AI model.
	GetName() string

	// GetAPIKey returns the API key for the AI model.
	GetAPIKey() string

	// GetCommand generates a command based on user input.
	GetCommand(prompt string) (*AIResponse, error)

	// SetContext sets the context for the AI model.
	SetContext(context string)

	// GetContext returns the context for the AI model.
	GetContext() string

	// SetCommandType sets the command type for the AI model.
	SetCommandType(commandType string)

	// GetCommandType returns the command type for the AI model.
	GetCommandType() string
}

// AIResponse holds the parsed command and suggested next steps.
type AIResponse struct {
	// Prompt is the user prompt that is sent to the AI.
	Prompt string

	// Command is the command that is suggested by the AI.
	Command string

	// Answer is the answer that is provided by the AI.
	Answer string

	// NextSteps are the suggested next steps that are provided by the AI.
	NextSteps []string
}
