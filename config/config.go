package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prompt-ops/pops/common"
)

// connections stores all the connection configurations in memory.
var connections []common.Connection

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

	var loadedConnections []common.Connection
	if err := json.NewDecoder(file).Decode(&loadedConnections); err != nil {
		return err
	}

	connections = loadedConnections
	return nil
}

// SaveConnection saves a new connection or updates an existing one based on the connection name.
func SaveConnection(conn common.Connection) error {
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
func GetConnectionByName(connectionName string) (common.Connection, error) {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return common.Connection{}, err
		}
	}

	for _, conn := range connections {
		if strings.EqualFold(conn.Name, connectionName) {
			return conn, nil
		}
	}

	return common.Connection{}, fmt.Errorf("connection with name '%s' does not exist", connectionName)
}

// GetAllConnections retrieves all stored connections.
func GetAllConnections() ([]common.Connection, error) {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return nil, err
		}
	}

	return connections, nil
}

// GetConnectionsByType retrieves connections filtered by their type.
func GetConnectionsByType(connectionType string) ([]common.Connection, error) {
	allConnections, err := GetAllConnections()
	if err != nil {
		return nil, err
	}

	var filteredConnections []common.Connection
	for _, conn := range allConnections {
		if strings.EqualFold(conn.Type.GetMainType(), connectionType) {
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

	var updatedConnections []common.Connection
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
	connections = []common.Connection{}

	if err := writeConnections(); err != nil {
		return err
	}

	return nil
}

func DeleteAllConnectionsByType(connectionType string) error {
	if connections == nil {
		if err := loadConnections(); err != nil {
			return err
		}
	}

	var updatedConnections []common.Connection
	for _, conn := range connections {
		if !strings.EqualFold(conn.Type.GetMainType(), connectionType) {
			updatedConnections = append(updatedConnections, conn)
		}
	}

	connections = updatedConnections

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
