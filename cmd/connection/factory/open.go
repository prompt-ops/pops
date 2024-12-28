package factory

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/cmd/connection/cloud"
	"github.com/prompt-ops/pops/cmd/connection/db"
	"github.com/prompt-ops/pops/cmd/connection/kubernetes"
	"github.com/prompt-ops/pops/common"
)

// GetOpenModel returns a new openModel based on the connection type
func GetOpenModel(connection common.Connection) (tea.Model, error) {
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
