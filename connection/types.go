package connection

// ConnectionInterface interface definition
type ConnectionInterface interface {
	CheckAuthentication() error
	InitialContext() error
	GetContext() string
	PrintContext() string
	GetCommand(prompt string) (string, error)
	ExecuteCommand(command string) ([]byte, error)
	Type() string
	SubType() string
	CommandType() string
}

// AvailableConnectionTypes returns a list of available connection types
func AvailableConnectionTypes() []string {
	return []string{
		Cloud,
		Database,
		Kubernetes,
	}
}

const (
	Cloud      string = "cloud"
	Database   string = "database"
	Kubernetes string = "kubernetes"
)
