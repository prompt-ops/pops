package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Connection represents a connection configuration.
type Connection struct {
	Type              string            `json:"type"`
	SubType           string            `json:"subType"`
	Name              string            `json:"name"`
	ConnectionDetails ConnectionDetails `json:"connectionDetails"`
}

// UnmarshalJSON implements custom JSON decoding for the Connection struct.
// It checks the 'Type' field, then decodes 'connectionDetails' into the correct struct.
// If connectionDetails is null in JSON, it remains nil.
func (c *Connection) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion when calling json.Unmarshal
	type alias Connection
	aux := &struct {
		ConnectionDetails json.RawMessage `json:"connectionDetails"`
		*alias
	}{
		alias: (*alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// If "connectionDetails" is null or empty, leave it nil
	if len(aux.ConnectionDetails) == 0 || string(aux.ConnectionDetails) == "null" {
		c.ConnectionDetails = nil
		return nil
	}

	// Decode into the appropriate detail struct based on "Type"
	switch c.Type {
	case "cloud":
		var details CloudConnectionDetails
		if err := json.Unmarshal(aux.ConnectionDetails, &details); err != nil {
			return err
		}
		c.ConnectionDetails = details

	case "kubernetes":
		var details KubernetesConnectionDetails
		if err := json.Unmarshal(aux.ConnectionDetails, &details); err != nil {
			return err
		}
		c.ConnectionDetails = details

	case "database":
		var details DatabaseConnectionDetails
		if err := json.Unmarshal(aux.ConnectionDetails, &details); err != nil {
			return err
		}
		c.ConnectionDetails = details

	default:
		// If it's some other type, you can either leave it as nil
		// or decode into a generic struct/map as needed.
		c.ConnectionDetails = nil
	}

	return nil
}

// ConnectionDetails is the interface that all connection detail structs will implement.
type ConnectionDetails interface {
	TypeName() string
}

// -------------------- Cloud --------------------
type CloudConnectionDetails struct {
	Provider string `json:"provider"`
}

func (c CloudConnectionDetails) TypeName() string {
	return "cloud"
}

// ------------------ Kubernetes -----------------
type KubernetesConnectionDetails struct {
	SelectedContext string `json:"selectedContext"`
}

func (k KubernetesConnectionDetails) TypeName() string {
	return "kubernetes"
}

// -------------------- Database -----------------
type DatabaseConnectionDetails struct {
	ConnectionString string `json:"connectionString"`
	Driver           string `json:"driver"`
}

func (d DatabaseConnectionDetails) TypeName() string {
	return "database"
}

// connections stores all the connection configurations in memory.
var connections []Connection

// connectionsConfigFilePath defines the path to the connections configuration file.
var connectionsConfigFilePath = getConfigFilePath("connections.json")

// getConfigFilePath constructs the absolute path for the configuration file.
func getConfigFilePath(filename string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filename
	}
	configDir := filepath.Join(homeDir, ".pops")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return filename
		}
	}
	return filepath.Join(configDir, filename)
}

// init loads existing connections from the configuration file into memory.
func init() {
	if err := loadConnections(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warning: Unable to load connections: %v\n", err)
	}
}

// loadConnections reads the connections from the JSON configuration file.
func loadConnections() error {
	file, err := os.Open(connectionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var loadedConnections []Connection
	if err := json.NewDecoder(file).Decode(&loadedConnections); err != nil {
		return err
	}

	connections = loadedConnections
	return nil
}

// SaveConnection saves a new connection or updates an existing one based on the connection name.
func SaveConnection(conn Connection) error {
	if connections == nil {
		if err := loadConnections(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	updated := false
	for i, existingConnection := range connections {
		if strings.EqualFold(existingConnection.Name, conn.Name) {
			connections[i] = conn
			updated = true
			break
		}
	}

	if !updated {
		connections = append(connections, conn)
	}

	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

// writeConnections writes the in-memory connections slice to the JSON configuration file.
func writeConnections() error {
	file, err := os.Create(connectionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(connections); err != nil {
		return err
	}

	return nil
}

// GetConnectionByName retrieves a connection by its name.
func GetConnectionByName(connectionName string) (Connection, error) {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return Connection{}, err
		}
	}

	for _, conn := range connections {
		if strings.EqualFold(conn.Name, connectionName) {
			return conn, nil
		}
	}

	return Connection{}, fmt.Errorf("connection with name '%s' does not exist", connectionName)
}

// GetAllConnections retrieves all stored connections.
func GetAllConnections() ([]Connection, error) {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return nil, err
		}
	}

	return connections, nil
}

// GetConnectionsByType retrieves connections filtered by their type.
func GetConnectionsByType(connectionType string) ([]Connection, error) {
	allConnections, err := GetAllConnections()
	if err != nil {
		return nil, err
	}

	var filteredConnections []Connection
	for _, conn := range allConnections {
		if strings.EqualFold(conn.Type, connectionType) {
			filteredConnections = append(filteredConnections, conn)
		}
	}

	return filteredConnections, nil
}

// DeleteConnectionByName removes a connection by its name.
func DeleteConnectionByName(connectionName string) error {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return err
		}
	}

	var updatedConnections []Connection
	found := false
	for _, conn := range connections {
		if strings.EqualFold(conn.Name, connectionName) {
			found = true
			continue
		}
		updatedConnections = append(updatedConnections, conn)
	}

	if !found {
		return fmt.Errorf("connection with name '%s' does not exist", connectionName)
	}

	connections = updatedConnections

	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

// DeleteAllConnections removes all stored connections.
func DeleteAllConnections() error {
	connections = []Connection{}

	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

// CheckIfNameExists checks if a connection with the given name already exists.
func CheckIfNameExists(name string) bool {
	if connections == nil {
		if err := loadConnections(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false
		}
	}

	for _, conn := range connections {
		if strings.EqualFold(conn.Name, name) {
			return true
		}
	}

	return false
}

// -------------------- Constructors --------------------

// NewCloudConnection creates a new cloud connection.
func NewCloudConnection(name, provider string) Connection {
	return Connection{
		Type:    "cloud",
		Name:    name,
		SubType: provider,
		ConnectionDetails: CloudConnectionDetails{
			Provider: provider,
		},
	}
}

// NewKubernetesConnection creates a new Kubernetes connection.
func NewKubernetesConnection(name, context string) Connection {
	return Connection{
		Type:    "kubernetes",
		Name:    name,
		SubType: context,
		ConnectionDetails: KubernetesConnectionDetails{
			SelectedContext: context,
		},
	}
}

// NewDatabaseConnection creates a new database connection.
func NewDatabaseConnection(name, driver, connectionString string) Connection {
	return Connection{
		Type:    "database",
		Name:    name,
		SubType: driver,
		ConnectionDetails: DatabaseConnectionDetails{
			ConnectionString: connectionString,
			// Driver is used to differentiate between different database types.
			Driver: driver,
		},
	}
}

// -------------------- Getters --------------------

// GetCloudConnectionDetails retrieves the CloudConnectionDetails from a Connection.
func GetCloudConnectionDetails(conn Connection) (CloudConnectionDetails, error) {
	if conn.Type != "cloud" {
		return CloudConnectionDetails{}, fmt.Errorf("connection is not of type 'cloud'")
	}
	details, ok := conn.ConnectionDetails.(CloudConnectionDetails)
	if !ok {
		return CloudConnectionDetails{}, fmt.Errorf("invalid connection details for 'cloud'")
	}
	return details, nil
}

// GetKubernetesConnectionDetails retrieves the KubernetesConnectionDetails from a Connection.
func GetKubernetesConnectionDetails(conn Connection) (KubernetesConnectionDetails, error) {
	if conn.Type != "kubernetes" {
		return KubernetesConnectionDetails{}, fmt.Errorf("connection is not of type 'kubernetes'")
	}
	details, ok := conn.ConnectionDetails.(KubernetesConnectionDetails)
	if !ok {
		return KubernetesConnectionDetails{}, fmt.Errorf("invalid connection details for 'kubernetes'")
	}
	return details, nil
}

// GetDatabaseConnectionDetails retrieves the DatabaseConnectionDetails from a Connection.
func GetDatabaseConnectionDetails(conn Connection) (DatabaseConnectionDetails, error) {
	if conn.Type != "database" {
		return DatabaseConnectionDetails{}, fmt.Errorf("connection is not of type 'database'")
	}
	details, ok := conn.ConnectionDetails.(DatabaseConnectionDetails)
	if !ok {
		return DatabaseConnectionDetails{}, fmt.Errorf("invalid connection details for 'database'")
	}
	return details, nil
}
