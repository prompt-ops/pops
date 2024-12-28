package common

import (
	"fmt"
	"strings"
)

// Factory function to get the right implementation based on type and subtype
func GetConnection(conn Connection) (ConnectionInterface, error) {
	switch conn.Type.GetMainType() {

	case ConnectionTypeCloud:
		switch strings.ToLower(conn.Type.GetSubtype()) {
		case "azure":
			return NewAzureConnection(&conn), nil
		default:
			return nil, fmt.Errorf("unsupported cloud subtype: %s", conn.Type.GetSubtype())
		}

	case ConnectionTypeKubernetes:
		return NewKubernetesConnectionImpl(&conn), nil

	case ConnectionTypeDatabase:
		switch strings.ToLower(conn.Type.GetSubtype()) {
		case "postgresql":
			return NewPostgreSQLConnection(&conn), nil
		default:
			return nil, fmt.Errorf("unsupported database subtype: %s", conn.Type.GetSubtype())
		}

	default:
		return nil, fmt.Errorf("[GetConnection] unsupported connection type: %s", conn.Type.GetSubtype())
	}
}
