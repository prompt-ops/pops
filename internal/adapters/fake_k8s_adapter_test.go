package adapters

import (
	"context"
	"strings"
	"testing"

	"github.com/prompt-ops/pops/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestFakeKubernetesAdapter_TestConnection(t *testing.T) {
	adapter := NewFakeKubernetesAdapter()

	tests := []struct {
		name        string
		connection  *domain.Connection
		expectError bool
	}{
		{
			name: "valid minikube cluster",
			connection: &domain.Connection{
				Name:    "test-k8s",
				Type:    domain.ConnectionTypeKubernetes,
				Subtype: "minikube",
			},
			expectError: false,
		},
		{
			name: "valid prod cluster",
			connection: &domain.Connection{
				Name:    "prod-k8s",
				Type:    domain.ConnectionTypeKubernetes,
				Subtype: "prod-cluster",
			},
			expectError: false,
		},
		{
			name: "invalid cluster",
			connection: &domain.Connection{
				Name:    "invalid-k8s",
				Type:    domain.ConnectionTypeKubernetes,
				Subtype: "non-existent",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := adapter.TestConnection(ctx, tt.connection)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFakeKubernetesAdapter_FetchContext(t *testing.T) {
	adapter := NewFakeKubernetesAdapter()

	tests := []struct {
		name             string
		connection       *domain.Connection
		expectError      bool
		expectedPods     int
		expectedServices int
	}{
		{
			name: "minikube cluster context",
			connection: &domain.Connection{
				Name:    "test-k8s",
				Type:    domain.ConnectionTypeKubernetes,
				Subtype: "minikube",
			},
			expectError:      false,
			expectedPods:     4,
			expectedServices: 3,
		},
		{
			name: "production cluster context",
			connection: &domain.Connection{
				Name:    "prod-k8s",
				Type:    domain.ConnectionTypeKubernetes,
				Subtype: "prod-cluster",
			},
			expectError:      false,
			expectedPods:     4,
			expectedServices: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			context, err := adapter.FetchContext(ctx, tt.connection)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, context)

				// Check that context contains expected information
				assert.Contains(t, context, "Cluster:")
				assert.Contains(t, context, "Nodes:")
				assert.Contains(t, context, "Pods:")
				assert.Contains(t, context, "Services:")
				assert.Contains(t, context, "Deployments:")

				// Count pods in context
				podLines := strings.Count(context, "pod:")
				assert.Equal(t, tt.expectedPods, podLines)

				// Count services in context
				serviceLines := strings.Count(context, "service:")
				assert.Equal(t, tt.expectedServices, serviceLines)
			}
		})
	}
}

func TestFakeKubernetesAdapter_ExecuteCommand(t *testing.T) {
	adapter := NewFakeKubernetesAdapter()
	connection := &domain.Connection{
		Name:    "test-k8s",
		Type:    domain.ConnectionTypeKubernetes,
		Subtype: "minikube",
	}

	tests := []struct {
		name            string
		command         string
		expectError     bool
		expectedContent []string
	}{
		{
			name:        "get pods command",
			command:     "kubectl get pods",
			expectError: false,
			expectedContent: []string{
				"NAME",
				"READY",
				"STATUS",
				"nginx-deployment-abc123",
				"postgres-7c9f8b",
			},
		},
		{
			name:        "get services command",
			command:     "kubectl get services",
			expectError: false,
			expectedContent: []string{
				"NAME",
				"TYPE",
				"nginx-service",
				"postgres-service",
				"ClusterIP",
			},
		},
		{
			name:        "get deployments command",
			command:     "kubectl get deployments",
			expectError: false,
			expectedContent: []string{
				"NAME",
				"READY",
				"nginx-deployment",
				"postgres",
			},
		},
		{
			name:        "get nodes command",
			command:     "kubectl get nodes",
			expectError: false,
			expectedContent: []string{
				"NAME",
				"STATUS",
				"minikube",
				"Ready",
			},
		},
		{
			name:        "unknown command",
			command:     "kubectl get secrets",
			expectError: false,
			expectedContent: []string{
				"Executed:",
				"kubectl get secrets",
				"Command completed successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := adapter.ExecuteCommand(ctx, connection, tt.command)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)

				for _, expected := range tt.expectedContent {
					assert.Contains(t, result, expected)
				}
			}
		})
	}
}

func TestFakeKubernetesAdapter_ClusterManagement(t *testing.T) {
	adapter := NewFakeKubernetesAdapter()

	// Test listing clusters
	clusters := adapter.ListClusters()
	assert.Contains(t, clusters, "minikube")
	assert.Contains(t, clusters, "prod-cluster")
	assert.Len(t, clusters, 2)

	// Test getting a specific cluster
	cluster := adapter.GetCluster("minikube")
	assert.NotNil(t, cluster)
	assert.Equal(t, "minikube", cluster.Name)
	assert.Len(t, cluster.Pods, 4)
	assert.Len(t, cluster.Services, 3)

	// Test adding a new cluster
	newCluster := &FakeCluster{
		Name:  "staging",
		Nodes: []string{"staging-node"},
		Pods: []FakePod{
			{Name: "test-pod", Namespace: "staging", Status: "Running", Node: "staging-node"},
		},
		Services: []FakeService{
			{Name: "test-service", Namespace: "staging", Type: "ClusterIP", Port: 8080},
		},
		Deployments: []FakeDeployment{
			{Name: "test-deployment", Namespace: "staging", Replicas: 1, Image: "test:latest"},
		},
	}

	adapter.AddCluster(newCluster)

	updatedClusters := adapter.ListClusters()
	assert.Len(t, updatedClusters, 3)
	assert.Contains(t, updatedClusters, "staging")

	retrievedCluster := adapter.GetCluster("staging")
	assert.NotNil(t, retrievedCluster)
	assert.Equal(t, "staging", retrievedCluster.Name)
}

func TestFakeKubernetesAdapter_ProductionCluster(t *testing.T) {
	adapter := NewFakeKubernetesAdapter()
	connection := &domain.Connection{
		Name:    "prod-k8s",
		Type:    domain.ConnectionTypeKubernetes,
		Subtype: "prod-cluster",
	}

	ctx := context.Background()

	// Test context contains production-specific data
	context, err := adapter.FetchContext(ctx, connection)
	assert.NoError(t, err)
	assert.Contains(t, context, "frontend-web")
	assert.Contains(t, context, "api-backend")
	assert.Contains(t, context, "production")
	assert.Contains(t, context, "node-1")
	assert.Contains(t, context, "node-2")
	assert.Contains(t, context, "node-3")

	// Test getting pods shows production pods
	result, err := adapter.ExecuteCommand(ctx, connection, "kubectl get pods")
	assert.NoError(t, err)
	assert.Contains(t, result, "frontend-web-abc123")
	assert.Contains(t, result, "api-backend-ghi789")
	assert.Contains(t, result, "database-primary")

	// Test getting services shows LoadBalancer type
	result, err = adapter.ExecuteCommand(ctx, connection, "kubectl get services")
	assert.NoError(t, err)
	assert.Contains(t, result, "frontend-service")
	assert.Contains(t, result, "LoadBalancer")
}
