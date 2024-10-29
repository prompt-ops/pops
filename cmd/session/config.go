package session

import (
	"encoding/json"
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

const configFilePath = "sessions.json"

func SaveSession(session Session) error {
	var sessions []Session

	// Read existing connections from the file
	file, err := os.Open(configFilePath)
	if err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&sessions); err != nil {
			return err
		}
	}

	sessions = append(sessions, session)

	file, err = os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(sessions)
}

func ListSessions() ([]Session, error) {
	var sessions []Session

	file, err := os.Open(configFilePath)
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
