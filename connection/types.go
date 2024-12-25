package connection

import (
	"fmt"

	config "github.com/prompt-ops/cli/config"
	cloud "github.com/prompt-ops/cli/connection/cloud"
	k8s "github.com/prompt-ops/cli/connection/kubernetes"
)

// PromptOpsConnection interface definition
type PromptOpsConnection interface {
	CheckAuthentication() error
	InitialContext() error
	GetContext() string
	PrintContext() string
	GetCommand(prompt string) (string, error)
	Type() string
	SubType() string
	CommandType() string
}

// Factory function to get the right implementation based on type and subtype
func GetConnection(conn config.Connection) (PromptOpsConnection, error) {
	switch conn.Type {
	case "cloud":
		switch conn.SubType {
		case "azure":
			return cloud.NewAzureConnection(conn), nil
		default:
			return nil, fmt.Errorf("unsupported cloud subtype: %s", conn.SubType)
		}
	case "kubernetes":
		return k8s.NewKubernetesConnection(conn), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.SubType)
	}
}
