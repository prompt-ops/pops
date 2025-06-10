package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prompt-ops/pops/internal/adapters"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]
	ctx := context.Background()

	// Setup services
	repo := NewMemoryConnectionRepository()
	k8sAdapter := adapters.NewFakeKubernetesAdapter()
	connService := services.NewConnectionService(repo, k8sAdapter)
	shellService := services.NewShellService(connService, &MockCommandService{}, &MockSessionService{})

	switch command {
	case "conn":
		handleConnectionCommands(ctx, connService, os.Args[2:])
	case "shell":
		handleShellCommands(ctx, shellService, os.Args[2:])
	case "clusters":
		handleClusterCommands(ctx, k8sAdapter)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
	}
}

func showUsage() {
	fmt.Println("🤖 PromptOps Manual Testing CLI")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/manual-test/main.go <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  conn create <name> <type> [subtype]  - Create a connection")
	fmt.Println("  conn list                           - List all connections")
	fmt.Println("  conn delete <name>                  - Delete a connection")
	fmt.Println("  shell context <name>                - Show connection context")
	fmt.Println("  clusters                            - List available fake clusters")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/manual-test/main.go conn create my-k8s kubernetes minikube")
	fmt.Println("  go run cmd/manual-test/main.go conn list")
	fmt.Println("  go run cmd/manual-test/main.go shell context my-k8s")
	fmt.Println("  go run cmd/manual-test/main.go clusters")
}

func handleConnectionCommands(ctx context.Context, connService *services.ConnectionService, args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Connection command required")
		return
	}

	switch args[0] {
	case "create":
		if len(args) < 3 {
			fmt.Println("❌ Usage: conn create <name> <type> [subtype]")
			return
		}
		
		name := args[1]
		connType := domain.ConnectionType(args[2])
		subtype := ""
		if len(args) > 3 {
			subtype = args[3]
		}

		req := services.CreateConnectionRequest{
			Name:    name,
			Type:    connType,
			Subtype: subtype,
		}

		conn, err := connService.CreateConnection(ctx, req)
		if err != nil {
			fmt.Printf("❌ Failed to create connection: %v\n", err)
			return
		}

		fmt.Printf("✅ Created connection '%s'\n", conn.Name)
		fmt.Printf("   Type: %s\n", conn.Type)
		fmt.Printf("   Subtype: %s\n", conn.Subtype)
		fmt.Printf("   Status: %s\n", conn.Status)
		fmt.Printf("   Context: %d characters\n", len(conn.Context))

	case "list":
		connections, err := connService.ListConnections(ctx, services.ConnectionFilter{})
		if err != nil {
			fmt.Printf("❌ Failed to list connections: %v\n", err)
			return
		}

		if len(connections) == 0 {
			fmt.Println("📭 No connections found")
			return
		}

		fmt.Printf("📋 Found %d connection(s):\n", len(connections))
		for i, conn := range connections {
			fmt.Printf("  %d. %s (%s/%s) - %s\n", i+1, conn.Name, conn.Type, conn.Subtype, conn.Status)
		}

	case "delete":
		if len(args) < 2 {
			fmt.Println("❌ Usage: conn delete <name>")
			return
		}

		name := args[1]
		err := connService.DeleteConnection(ctx, name)
		if err != nil {
			fmt.Printf("❌ Failed to delete connection: %v\n", err)
			return
		}

		fmt.Printf("✅ Deleted connection '%s'\n", name)

	default:
		fmt.Printf("❌ Unknown connection command: %s\n", args[0])
	}
}

func handleShellCommands(ctx context.Context, shellService *services.ShellService, args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Shell command required")
		return
	}

	switch args[0] {
	case "context":
		if len(args) < 2 {
			fmt.Println("❌ Usage: shell context <connection-name>")
			return
		}

		connectionName := args[1]
		display, err := shellService.ShowContext(ctx, connectionName)
		if err != nil {
			fmt.Printf("❌ Failed to show context: %v\n", err)
			return
		}

		fmt.Printf("🖥️  %s\n", display.Title)
		fmt.Printf("📝 %s\n", display.Description)
		fmt.Printf("🔗 Connection: %s (%s)\n", display.ConnectionName, display.ConnectionType)
		fmt.Println("")

		if len(display.Items) == 0 {
			fmt.Println("📭 No resources found")
			return
		}

		fmt.Printf("📦 Resources (%d):\n", len(display.Items))
		for _, item := range display.Items {
			prefix := ""
			if item.Parent != "" {
				prefix = "  └─ "
			}
			fmt.Printf("  %s%s %s (%s)\n", prefix, item.Icon, item.Name, item.Type)
		}

	default:
		fmt.Printf("❌ Unknown shell command: %s\n", args[0])
	}
}

func handleClusterCommands(ctx context.Context, adapter *adapters.FakeKubernetesAdapter) {
	clusters := adapter.ListClusters()
	
	fmt.Printf("🎯 Available Fake Clusters (%d):\n", len(clusters))
	for i, clusterName := range clusters {
		cluster := adapter.GetCluster(clusterName)
		fmt.Printf("  %d. %s\n", i+1, clusterName)
		fmt.Printf("     Nodes: %d\n", len(cluster.Nodes))
		fmt.Printf("     Pods: %d\n", len(cluster.Pods))
		fmt.Printf("     Services: %d\n", len(cluster.Services))
		fmt.Printf("     Deployments: %d\n", len(cluster.Deployments))
		
		if i < len(clusters)-1 {
			fmt.Println()
		}
	}
	
	fmt.Println("\nTo test a cluster:")
	fmt.Println("  go run cmd/manual-test/main.go conn create my-cluster kubernetes " + clusters[0])
	fmt.Println("  go run cmd/manual-test/main.go shell context my-cluster")
}

// Copy the mock implementations from test-harness
type MemoryConnectionRepository struct {
	connections map[string]*domain.Connection
}

func NewMemoryConnectionRepository() *MemoryConnectionRepository {
	return &MemoryConnectionRepository{
		connections: make(map[string]*domain.Connection),
	}
}

func (r *MemoryConnectionRepository) Save(ctx context.Context, conn *domain.Connection) error {
	r.connections[conn.Name] = conn
	return nil
}

func (r *MemoryConnectionRepository) FindByName(ctx context.Context, name string) (*domain.Connection, error) {
	conn, exists := r.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", name)
	}
	return conn, nil
}

func (r *MemoryConnectionRepository) FindAll(ctx context.Context) ([]*domain.Connection, error) {
	var connections []*domain.Connection
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *MemoryConnectionRepository) Delete(ctx context.Context, name string) error {
	delete(r.connections, name)
	return nil
}

type MockCommandService struct{}

func (m *MockCommandService) GenerateCommand(ctx context.Context, req services.GenerateCommandRequest) (*domain.Command, error) {
	return &domain.Command{
		ID:          "cmd-1",
		Type:        domain.CommandTypeQuery,
		Content:     "kubectl get pods",
		Description: "List all pods",
		Safety:      domain.SafetyLevelSafe,
	}, nil
}

func (m *MockCommandService) GenerateAnswer(ctx context.Context, req services.GenerateAnswerRequest) (string, error) {
	return "This is a mock answer about your infrastructure.", nil
}

func (m *MockCommandService) ExecuteCommand(ctx context.Context, commandID domain.CommandID, dryRun bool) (*domain.ExecutionResult, error) {
	return &domain.ExecutionResult{
		CommandID: commandID,
		Status:    domain.ExecutionStatusCompleted,
		Output:    "Mock command executed successfully",
	}, nil
}

func (m *MockCommandService) GetCommandHistory(ctx context.Context, connectionName string) ([]*domain.Command, error) {
	return []*domain.Command{}, nil
}

type MockSessionService struct{}

func (m *MockSessionService) CreateSession(ctx context.Context, connectionID domain.ConnectionID) (*domain.Session, error) {
	return &domain.Session{
		ID:           "session-1",
		ConnectionID: connectionID,
		State:        domain.SessionStateActive,
	}, nil
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error) {
	return &domain.Session{
		ID:    sessionID,
		State: domain.SessionStateActive,
	}, nil
}

func (m *MockSessionService) AddInteraction(ctx context.Context, sessionID domain.SessionID, interaction domain.Interaction) error {
	return nil
}

func (m *MockSessionService) UpdateSessionState(ctx context.Context, sessionID domain.SessionID, state domain.SessionState) error {
	return nil
}

func (m *MockSessionService) CloseSession(ctx context.Context, sessionID domain.SessionID) error {
	return nil
}

func (m *MockSessionService) GetActiveSessions(ctx context.Context) ([]*domain.Session, error) {
	return []*domain.Session{}, nil
}