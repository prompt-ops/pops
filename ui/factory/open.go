package factory

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/config"
	"github.com/prompt-ops/pops/ui/cloud"
	"github.com/prompt-ops/pops/ui/kubernetes"
)

// GetOpenModel returns a new openModel based on the connection type
func GetOpenModel(connection config.Connection) (tea.Model, error) {
	switch connection.Type {
	case "cloud":
		return cloud.NewCloudOpenModel(), nil
	case "kubernetes":
		return kubernetes.NewKubernetesOpenModel(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connection.Type)
	}
}
