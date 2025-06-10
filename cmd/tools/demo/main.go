package main

import (
	"context"
	"fmt"

	"github.com/prompt-ops/pops/internal/adapters"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
)

func main() {
	fmt.Println("🚀 PromptOps Interactive Demo")
	fmt.Println("=============================")

	ctx := context.Background()

	// Setup services (persistent in this session)
	repo := NewMemoryConnectionRepository()
	k8sAdapter := adapters.NewFakeKubernetesAdapter()
	connService := services.NewConnectionService(repo, k8sAdapter)
	shellService := services.NewShellService(connService, &MockCommandService{}, &MockSessionService{})

	// Demo workflow
	fmt.Println("\n🎬 Demo: Complete PromptOps Workflow")
	
	// Step 1: List available clusters
	fmt.Println("\n1️⃣ Available Clusters:")
	clusters := k8sAdapter.ListClusters()
	for i, cluster := range clusters {
		fmt.Printf("   %d. %s\n", i+1, cluster)
	}

	// Step 2: Create connections
	fmt.Println("\n2️⃣ Creating Connections:")
	connections := []struct{name, subtype string}{
		{"dev-minikube", "minikube"},
		{"production", "prod-cluster"},
	}

	for _, conn := range connections {
		req := services.CreateConnectionRequest{
			Name:    conn.name,
			Type:    domain.ConnectionTypeKubernetes,
			Subtype: conn.subtype,
		}

		result, err := connService.CreateConnection(ctx, req)
		if err != nil {
			fmt.Printf("   ❌ Failed to create %s: %v\n", conn.name, err)
			continue
		}
		fmt.Printf("   ✅ %s (%s) - %s\n", result.Name, result.Subtype, result.Status)
	}

	// Step 3: List connections
	fmt.Println("\n3️⃣ Active Connections:")
	allConns, err := connService.ListConnections(ctx, services.ConnectionFilter{})
	if err != nil {
		fmt.Printf("   ❌ Failed to list: %v\n", err)
	} else {
		for i, conn := range allConns {
			fmt.Printf("   %d. %s (%s/%s)\n", i+1, conn.Name, conn.Type, conn.Subtype)
		}
	}

	// Step 4: Show context for each connection
	fmt.Println("\n4️⃣ Connection Contexts:")
	for _, conn := range connections {
		fmt.Printf("\n📊 %s Context:\n", conn.name)
		
		display, err := shellService.ShowContext(ctx, conn.name)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
			continue
		}

		fmt.Printf("   %s\n", display.Title)
		fmt.Printf("   📝 %s\n", display.Description)
		
		// Group resources by type
		resourceGroups := make(map[string][]services.ContextItem)
		for _, item := range display.Items {
			resourceGroups[item.Type] = append(resourceGroups[item.Type], item)
		}

		for resourceType, items := range resourceGroups {
			if len(items) > 3 {
				fmt.Printf("   📦 %s: %d items (%s, %s, %s, ...)\n", 
					resourceType, len(items), items[0].Name, items[1].Name, items[2].Name)
			} else {
				fmt.Printf("   📦 %s: ", resourceType)
				for i, item := range items {
					if i > 0 { fmt.Print(", ") }
					fmt.Printf("%s", item.Name)
				}
				fmt.Println()
			}
		}
	}

	// Step 5: Simulate command execution
	fmt.Println("\n5️⃣ Command Simulation:")
	testCommands := map[string][]string{
		"dev-minikube": {"kubectl get pods", "kubectl get services"},
		"production": {"kubectl get pods", "kubectl get nodes"},
	}

	for connName, commands := range testCommands {
		fmt.Printf("\n⚡ %s commands:\n", connName)
		
		// Get the connection
		conn, err := connService.GetConnection(ctx, connName)
		if err != nil {
			fmt.Printf("   ❌ Connection not found: %v\n", err)
			continue
		}

		for _, cmd := range commands {
			result, err := k8sAdapter.ExecuteCommand(ctx, conn, cmd)
			if err != nil {
				fmt.Printf("   ❌ %s: %v\n", cmd, err)
			} else {
				lines := len(result)
				preview := result
				if lines > 150 {
					preview = result[:150] + "..."
				}
				fmt.Printf("   ✅ %s\n", cmd)
				fmt.Printf("      %s\n", preview)
			}
		}
	}

	// Step 6: Show difference between environments
	fmt.Println("\n6️⃣ Environment Comparison:")
	devConn, _ := connService.GetConnection(ctx, "dev-minikube")
	prodConn, _ := connService.GetConnection(ctx, "production")

	fmt.Println("   Development (minikube):")
	devResult, _ := k8sAdapter.ExecuteCommand(ctx, devConn, "kubectl get nodes")
	fmt.Printf("   %s\n", devResult)

	fmt.Println("   Production (multi-node):")
	prodResult, _ := k8sAdapter.ExecuteCommand(ctx, prodConn, "kubectl get nodes")
	fmt.Printf("   %s\n", prodResult)

	fmt.Println("\n🎉 Demo Complete!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("   ✅ Clean service-based architecture")
	fmt.Println("   ✅ Multiple cluster support")
	fmt.Println("   ✅ Rich context display with emojis")
	fmt.Println("   ✅ Command execution simulation")
	fmt.Println("   ✅ Environment-specific behavior")
	fmt.Println("   ✅ Proper error handling")
	fmt.Println("   ✅ Type-safe domain models")
}

// Minimal repository implementation for demo
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

// Mock services
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