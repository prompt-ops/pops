package factory

import (
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/cmd/pops/app/conn/cloud"
	"github.com/prompt-ops/pops/cmd/pops/app/conn/db"
	"github.com/prompt-ops/pops/cmd/pops/app/conn/k8s"
	"github.com/prompt-ops/pops/pkg/conn"

	tea "github.com/charmbracelet/bubbletea"
)

// GetOpenModel returns a new openModel based on the connection type
func GetOpenModel(connection conn.Connection) (tea.Model, error) {
	switch strings.ToLower(connection.Type.GetMainType()) {
	case "cloud":
		return cloud.NewOpenModel(), nil
	case "kubernetes":
		return k8s.NewOpenModel(), nil
	case "database":
		return db.NewOpenModel(), nil
	default:
		return nil, fmt.Errorf("[GetOpenModel] unsupported connection type: %s", connection.Type)
	}
}
