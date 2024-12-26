package factory

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
)

// GetCreateModel returns a new createModel based on the connection type
func GetCreateModel(connectionType string) (tea.Model, error) {
	switch connectionType {
	case "cloud":
		return cloud.NewCreateModel(), nil
	case "kubernetes":
		return kubernetes.NewCreateModel(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connectionType)
	}
}
