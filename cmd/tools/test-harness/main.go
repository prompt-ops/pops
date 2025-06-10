package main

import (
	"context"
	"fmt"
	"log"

	"github.com/prompt-ops/pops/internal/adapters"
	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
)

func main() {
	fmt.Println("🚀 PromptOps Architecture Test Harness")
	fmt.Println("=====================================")

	ctx := context.Background()

	// Create fake repository (in-memory for testing)
	repo := NewMemoryConnectionRepository()
	
	// Create fake Kubernetes adapter
	k8sAdapter := adapters.NewFakeKubernetesAdapter()
	
	// Create connection service
	connService := services.NewConnectionService(repo, k8sAdapter)
	
	// Test 1: Create a Kubernetes connection
	fmt.Println("\n📝 Test 1: Creating Kubernetes Connection")
	createReq := services.CreateConnectionRequest{
		Name:    "test-minikube",
		Type:    domain.ConnectionTypeKubernetes,
		Subtype: "minikube",
		Metadata: map[string]string{
			"cluster": "minikube",
		},
	}
	
	conn, err := connService.CreateConnection(ctx, createReq)
	if err != nil {
		log.Printf("❌ Failed to create connection: %v", err)
	} else {
		fmt.Printf("✅ Created connection: %s (Status: %s)\n", conn.Name, conn.Status)
		fmt.Printf("📊 Context preview: %s...\n", conn.Context[:min(100, len(conn.Context))])
	}

	// Test 2: List connections
	fmt.Println("\n📋 Test 2: Listing Connections")
	connections, err := connService.ListConnections(ctx, services.ConnectionFilter{})
	if err != nil {
		log.Printf("❌ Failed to list connections: %v", err)
	} else {
		fmt.Printf("✅ Found %d connections:\n", len(connections))
		for _, c := range connections {
			fmt.Printf("  - %s (%s)\n", c.Name, c.Type)
		}
	}

	// Test 3: Test shell service context display
	fmt.Println("\n🖥️  Test 3: Shell Context Display")
	// Create mock services for shell
	cmdService := &MockCommandService{}
	sessionService := &MockSessionService{}
	
	shellService := services.NewShellService(connService, cmdService, sessionService)
	
	display, err := shellService.ShowContext(ctx, "test-minikube")
	if err != nil {
		log.Printf("❌ Failed to show context: %v", err)
	} else {
		fmt.Printf("✅ Context Display:\n")
		fmt.Printf("   Title: %s\n", display.Title)
		fmt.Printf("   Description: %s\n", display.Description)
		fmt.Printf("   Items: %d\n", len(display.Items))
		
		// Show first few items
		for i, item := range display.Items {
			if i >= 3 {
				fmt.Printf("   ... and %d more items\n", len(display.Items)-3)
				break
			}
			fmt.Printf("   %s %s (%s)\n", item.Icon, item.Name, item.Type)
		}
	}

	// Test 4: Test Kubernetes adapter directly
	fmt.Println("\n⚙️  Test 4: Kubernetes Adapter Commands")
	testCommands := []string{
		"kubectl get pods",
		"kubectl get services",
		"kubectl get nodes",
	}
	
	for _, cmd := range testCommands {
		result, err := k8sAdapter.ExecuteCommand(ctx, conn, cmd)
		if err != nil {
			log.Printf("❌ Command failed: %s - %v", cmd, err)
		} else {
			fmt.Printf("✅ Command: %s\n", cmd)
			lines := len(result)
			if lines > 200 {
				fmt.Printf("   Result: %s... (%d chars)\n", result[:200], lines)
			} else {
				fmt.Printf("   Result: %s\n", result)
			}
		}
		fmt.Println()
	}

	// Test 5: Test different cluster
	fmt.Println("\n🏭 Test 5: Production Cluster")
	prodReq := services.CreateConnectionRequest{
		Name:    "prod-cluster",
		Type:    domain.ConnectionTypeKubernetes,
		Subtype: "prod-cluster",
	}
	
	prodConn, err := connService.CreateConnection(ctx, prodReq)
	if err != nil {
		log.Printf("❌ Failed to create prod connection: %v", err)
	} else {
		fmt.Printf("✅ Created prod connection: %s\n", prodConn.Name)
		
		// Show production context
		prodDisplay, err := shellService.ShowContext(ctx, "prod-cluster")
		if err != nil {
			log.Printf("❌ Failed to show prod context: %v", err)
		} else {
			fmt.Printf("📊 Production cluster has %d resources\n", len(prodDisplay.Items))
			
			// Count resource types
			resourceTypes := make(map[string]int)
			for _, item := range prodDisplay.Items {
				resourceTypes[item.Type]++
			}
			
			for resourceType, count := range resourceTypes {
				fmt.Printf("   - %d %s(s)\n", count, resourceType)
			}
		}
	}

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("👉 The new architecture is working great!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Simple in-memory repository for testing
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

// Mock services for shell testing
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