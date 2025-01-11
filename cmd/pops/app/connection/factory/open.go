package factory

import (
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/cmd/pops/app/connection/cloud"
	"github.com/prompt-ops/pops/cmd/pops/app/connection/db"
	"github.com/prompt-ops/pops/cmd/pops/app/connection/kubernetes"
	"github.com/prompt-ops/pops/pkg/connection"

	tea "github.com/charmbracelet/bubbletea"
)

// GetOpenModel returns a new openModel based on the connection type
func GetOpenModel(connection connection.Connection) (tea.Model, error) {
	switch strings.ToLower(connection.Type.GetMainType()) {
	case "cloud":
		return cloud.NewOpenModel(), nil
	case "kubernetes":
		return kubernetes.NewOpenModel(), nil
	case "database":
		return db.NewOpenModel(), nil
	default:
		return nil, fmt.Errorf("[GetOpenModel] unsupported connection type: %s", connection.Type)
	}
}
