package factory

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
	"github.com/prompt-ops/pops/config"
)

// GetOpenModel returns a new openModel based on the connection type
func GetOpenModel(connection config.Connection) (tea.Model, error) {
	switch connection.Type {
	case "cloud":
		return cloud.NewOpenModel(), nil
	case "kubernetes":
		return kubernetes.NewOpenModel(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connection.Type)
	}
}
