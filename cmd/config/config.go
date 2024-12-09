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
		if existingSession.Name == session.Name {
			return fmt.Errorf("another session with the same name already exists")
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

func DeleteSessionByName(sessionName string) error {
	existingSessions, err := ListSessions()
	if err != nil {
		return err
	}

	var updatedSessions []Session
	for _, session := range existingSessions {
		if session.Name == sessionName {
			continue
		}
		updatedSessions = append(updatedSessions, session)
	}

	file, err := os.Create(sessionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(updatedSessions)
}

func DeleteAllSessions() error {
	file, err := os.Create(sessionsConfigFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	emptyList := []Session{}
	if err := json.NewEncoder(file).Encode(emptyList); err != nil {
		return err
	}

	return nil
}
