package connection

import (
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/config"
	"github.com/prompt-ops/pops/connection/cloud"
	"github.com/prompt-ops/pops/connection/db"
	"github.com/prompt-ops/pops/connection/kubernetes"
)

// Factory function to get the right implementation based on type and subtype
func GetConnection(conn config.Connection) (ConnectionInterface, error) {
	switch strings.ToLower(conn.Type) {

	case "cloud":
		switch strings.ToLower(conn.SubType) {
		case "azure":
			return cloud.NewAzureConnection(conn), nil
		default:
			return nil, fmt.Errorf("unsupported cloud subtype: %s", conn.SubType)
		}

	case "kubernetes":
		return kubernetes.NewKubernetesConnection(conn), nil

	case "database":
		switch strings.ToLower(conn.SubType) {
		case "postgresql":
			return db.NewPostgreSQLConnection(&conn), nil
		default:
			return nil, fmt.Errorf("unsupported database subtype: %s", conn.SubType)
		}

	default:
		return nil, fmt.Errorf("[GetConnection] unsupported connection type: %s", conn.SubType)
	}
}
