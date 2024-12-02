package connection

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

func SaveSession(session Session) error {
	var sessions []Session

	// Read existing connections from the file
	file, err := os.Open(sessionsConfigFilePath)
	if err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&sessions); err != nil {
			return err
		}
	}

	// Check if a session with the same name, connection name, and connection type already exists
	for _, existingSession := range sessions {
		if existingSession.Name == session.Name &&
			existingSession.Connection.Name == session.Connection.Name &&
			existingSession.Connection.Type == session.Connection.Type {
			return fmt.Errorf("session with the same name, connection name, and connection type already exists")
		}
	}

	sessions = append(sessions, session)

	file, err = os.Create(sessionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(sessions)
}

func ListSessions() ([]Session, error) {
	var sessions []Session

	file, err := os.Open(sessionsConfigFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func ListSessionsByConnection(connectionName string) ([]Session, error) {
	sessions, err := ListSessions()
	if err != nil {
		return nil, err
	}

	var filteredSessions []Session
	for _, session := range sessions {
		if session.Connection.Name == connectionName {
			filteredSessions = append(filteredSessions, session)
		}
	}

	return filteredSessions, nil
}
