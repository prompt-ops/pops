package conn

import (
	"encoding/json"
	"fmt"
)

var (
	ConnectionTypeCloud      = "Cloud"
	ConnectionTypeDatabase   = "Database"
	ConnectionTypeKubernetes = "Kubernetes"
)

// AvailableConnectionTypes is a list of available connection types.
func AvailableConnectionTypes() []string {
	return []string{
		ConnectionTypeCloud,
		ConnectionTypeDatabase,
		ConnectionTypeKubernetes,
	}
}

type ConnectionType interface {
	// GetMainType returns the main type of the connection.
	// Example: "database", "cloud", "kubernetes".
	GetMainType() string

	// GetSubtype returns the subtype of the connection.
	// Example: "postgres", "mysql", "aws", "gcp", "azure".
	// Can be empty if there is no subtype.
	GetSubtype() string
}

type ConnectionDetails interface {
	// GetDriver returns the driver name for the connection.
	// Example: "postgres", "mysql", "mongodb".
	// Can be empty if there is no driver.
	GetDriver() string
}

type Connection struct {
	// Name of the connection.
	Name string `json:"name"`

	// Type of the connection.
	// Example: "database", "cloud", "kubernetes".
	Type ConnectionType `json:"type"`

	// Details of the connection.
	// Can include different details based on the connection type.
	// Like database connection string, cloud credentials, etc.
	Details ConnectionDetails `json:"details"`
}

// UnmarshalJSON implements custom JSON decoding for the Connection struct.
func (c *Connection) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion when calling json.Unmarshal
	type alias Connection
	aux := &struct {
		Type    json.RawMessage `json:"type"`
		Details json.RawMessage `json:"details"`
		*alias
	}{
		alias: (*alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal the Type field based on the main type
	var mainType struct {
		MainType string `json:"mainType"`
	}
	if err := json.Unmarshal(aux.Type, &mainType); err != nil {
		return err
	}

	switch mainType.MainType {
	case ConnectionTypeDatabase:
		var dbType DatabaseConnectionType
		if err := json.Unmarshal(aux.Type, &dbType); err != nil {
			return err
		}
		c.Type = dbType

		var dbDetails DatabaseConnectionDetails
		if err := json.Unmarshal(aux.Details, &dbDetails); err != nil {
			return err
		}
		c.Details = dbDetails

	case ConnectionTypeKubernetes:
		var k8sType KubernetesConnectionType
		if err := json.Unmarshal(aux.Type, &k8sType); err != nil {
			return err
		}
		c.Type = k8sType

		var k8sDetails KubernetesConnectionDetails
		if err := json.Unmarshal(aux.Details, &k8sDetails); err != nil {
			return err
		}
		c.Details = k8sDetails

	case ConnectionTypeCloud:
		var cloudType CloudConnectionType
		if err := json.Unmarshal(aux.Type, &cloudType); err != nil {
			return err
		}
		c.Type = cloudType

		var cloudDetails CloudConnectionDetails
		if err := json.Unmarshal(aux.Details, &cloudDetails); err != nil {
			return err
		}
		c.Details = cloudDetails

	default:
		return fmt.Errorf("unknown main type: %s", mainType.MainType)
	}

	return nil
}

type ConnectionInterface interface {
	GetConnection() Connection
	CheckAuthentication() error

	// SetContext gets the necessary information for the connection.
	// For example, for a database connection, it can get the list of tables and columns.
	// For a cloud connection, it can get the list of resources.
	// For a kubernetes connection, it can get the list of deployments, services, etc.
	// This information will be sent to the AI model which will use it to generate the queries/commands.
	SetContext() error

	// GetContext returns the information set by the SetContext method.
	// This information will be sent to the AI model which will use it to generate the queries/commands.
	GetContext() string

	// GetFormattedContext returns the formatted context for the AI model.
	GetFormattedContext() (string, error)

	// ExecuteCommand executes the given command and returns the output as byte array.
	ExecuteCommand(command string) ([]byte, error)

	// FormatResultAsTable formats the result as a table.
	FormatResultAsTable(result []byte) (string, error)

	// GetCommand gets the command from AI using context and the user prompt.
	GetCommand(prompt string) (string, error)

	// GetAnswer gets the answer from AI using context and the user prompt.
	GetAnswer(prompt string) (string, error)

	// CommandType returns the type of the command.
	// Example: "psql", "az", "kubectl".
	CommandType() string
}
