package connection

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Connection struct {
	Type     string    `json:"type"`
	SubType  string    `json:"subType"`
	Name     string    `json:"name"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	Name       string     `json:"name"`
	Connection Connection `json:"connection"`
}

const sessionsConfigFilePath = "sessions.json"
const connectionsConfigFilePath = "connections.json"

// TODO: This may need to overwrite the existing connection if the name is the same.
// We can check uniqueness by another field like the connection string.
func SaveConnection(conn Connection) error {
	var connections []Connection

	// Read existing connections from the file
	file, err := os.Open(connectionsConfigFilePath)
	if err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&connections); err != nil {
			return err
		}
	}

	// Check if a connection with the same name
	for _, existingConnection := range connections {
		if existingConnection.Name == conn.Name {
			return fmt.Errorf("connection with the same name already exists")
		}
	}

	// Append the new connection
	connections = append(connections, conn)

	// Write the updated connections back to the file
	file, err = os.Create(connectionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(connections)
}

func ListConnections() ([]Connection, error) {
	var connections []Connection

	file, err := os.Open(connectionsConfigFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&connections); err != nil {
		return nil, err
	}

	return connections, nil
}

// Make this more efficient.
func DeleteConnectionByName(connectionName string) error {
	existingConnections, err := ListConnections()
	if err != nil {
		return err
	}

	var updatedConnections []Connection
	for _, conn := range existingConnections {
		if conn.Name == connectionName {
			continue
		}
		updatedConnections = append(updatedConnections, conn)
	}

	file, err := os.Create(connectionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(updatedConnections)
}

func DeleteAllConnections() error {
	file, err := os.Create(connectionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	emptyList := []Connection{}
	if err := json.NewEncoder(file).Encode(emptyList); err != nil {
		return err
	}

	return nil
}

func CheckIfNameExists(name string) bool {
	connections, err := ListConnections()
	if err != nil {
		return false
	}

	for _, conn := range connections {
		if strings.ToLower(conn.Name) == strings.ToLower(name) {
			return true
		}
	}

	return false
}
