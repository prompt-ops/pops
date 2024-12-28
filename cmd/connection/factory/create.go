package factory

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/db"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
)

// GetCreateModel returns a new createModel based on the connection type
func GetCreateModel(connectionType string) (tea.Model, error) {
	switch strings.ToLower(connectionType) {
	case "cloud":
		return cloud.NewCreateModel(), nil
	case "kubernetes":
		return kubernetes.NewCreateModel(), nil
	case "database":
		return db.NewCreateModel(), nil
	default:
		return nil, fmt.Errorf("[GetCreateModel] unsupported connection type: %s", connectionType)
	}
}
