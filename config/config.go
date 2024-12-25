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
	Type    string `json:"type"`
	SubType string `json:"subType"`
	Name    string `json:"name"`
}

// connections stores all the connection configurations in memory.
var connections []Connection

// connectionsConfigFilePath defines the path to the connections configuration file.
// It's recommended to store configuration files in the user's home directory.
var connectionsConfigFilePath = getConfigFilePath("connections.json")

// getConfigFilePath constructs the absolute path for the configuration file.
// It ensures that the configuration directory exists.
func getConfigFilePath(filename string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		return filename
	}
	configDir := filepath.Join(homeDir, ".pops")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			// If unable to create config directory, fallback to current directory
			return filename
		}
	}
	return filepath.Join(configDir, filename)
}

// init loads existing connections from the configuration file into memory.
// This ensures that subsequent operations work with the in-memory data.
func init() {
	if err := loadConnections(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warning: Unable to load connections: %v\n", err)
	}
}

// loadConnections reads the connections from the JSON configuration file
// and loads them into the in-memory slice.
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
// If a connection with the same name exists, it will be overwritten with the new details.
func SaveConnection(conn Connection) error {
	// Load existing connections if not already loaded
	if connections == nil {
		if err := loadConnections(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	// Check if a connection with the same name exists
	updated := false
	for i, existingConnection := range connections {
		if strings.EqualFold(existingConnection.Name, conn.Name) {
			connections[i] = conn
			updated = true
			break
		}
	}

	if !updated {
		// Append the new connection if it doesn't exist
		connections = append(connections, conn)
	}

	// Write the updated connections back to the file
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
	encoder.SetIndent("", "  ") // Optional: for pretty-printing
	if err := encoder.Encode(connections); err != nil {
		return err
	}

	return nil
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

// GetConnectionsByType retrieves connections filtered by their type (e.g., "cloud", "kubernetes").
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

	if len(filteredConnections) == 0 {
		return nil, fmt.Errorf("no connections found for type '%s'", connectionType)
	}

	return filteredConnections, nil
}

// DeleteConnectionByName removes a connection by its name.
// If the connection does not exist, it returns an error.
func DeleteConnectionByName(connectionName string) error {
	// Load connections if not already loaded
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
			continue // Skip the connection to be deleted
		}
		updatedConnections = append(updatedConnections, conn)
	}

	if !found {
		return fmt.Errorf("connection with name '%s' does not exist", connectionName)
	}

	connections = updatedConnections

	// Write the updated connections back to the file
	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

// DeleteAllConnections removes all stored connections.
// Use with caution as this operation is irreversible.
func DeleteAllConnections() error {
	connections = []Connection{}

	// Write the empty connections list back to the file
	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

// CheckIfNameExists checks if a connection with the given name already exists.
// It performs a case-insensitive comparison.
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
