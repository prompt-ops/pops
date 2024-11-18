package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type Connection struct {
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	Name string `json:"name"`
	// Find a way to keep history of commands and responses.
	// This is a placeholder for now.
}

const configFilePath = "connections.json"

// TODO: This may need to overwrite the existing connection if the name is the same.
// Should return an error if the connection name for the same type already exists.
func SaveConnection(conn Connection) error {
	var connections []Connection

	// Read existing connections from the file
	file, err := os.Open(configFilePath)
	if err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&connections); err != nil {
			return err
		}
	}

	// Check if the connection name for the same type already exists
	for _, c := range connections {
		if c.Type == conn.Type && c.Name == conn.Name {
			return fmt.Errorf("connection name '%s' already exists for connection type '%s'", conn.Name, conn.Type)
		}
	}

	// Append the new connection
	connections = append(connections, conn)

	// Write the updated connections back to the file
	file, err = os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(connections)
}

func ListConnections() ([]Connection, error) {
	var connections []Connection

	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&connections); err != nil {
		return nil, err
	}

	return connections, nil
}
