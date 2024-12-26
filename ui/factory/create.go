package factory

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/ui/cloud"
	"github.com/prompt-ops/pops/ui/kubernetes"
)

// GetCreateModel returns a new createModel based on the connection type
func GetCreateModel(connectionType string) (tea.Model, error) {
	switch connectionType {
	case "cloud":
		return cloud.NewCloudCreateModel(), nil
	case "kubernetes":
		return kubernetes.NewKubernetesCreateModel(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connectionType)
	}
}
