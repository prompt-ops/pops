package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/internal/domain"
)

type FakeKubernetesAdapter struct {
	clusters map[string]*FakeCluster
}

type FakeCluster struct {
	Name        string
	Nodes       []string
	Pods        []FakePod
	Services    []FakeService
	Deployments []FakeDeployment
}

type FakePod struct {
	Name      string
	Namespace string
	Status    string
	Node      string
}

type FakeService struct {
	Name      string
	Namespace string
	Type      string
	Port      int
}

type FakeDeployment struct {
	Name      string
	Namespace string
	Replicas  int
	Image     string
}

func NewFakeKubernetesAdapter() *FakeKubernetesAdapter {
	adapter := &FakeKubernetesAdapter{
		clusters: make(map[string]*FakeCluster),
	}

	// Create a default fake cluster
	adapter.clusters["minikube"] = &FakeCluster{
		Name:  "minikube",
		Nodes: []string{"minikube"},
		Pods: []FakePod{
			{Name: "nginx-deployment-abc123", Namespace: "default", Status: "Running", Node: "minikube"},
			{Name: "nginx-deployment-def456", Namespace: "default", Status: "Running", Node: "minikube"},
			{Name: "postgres-7c9f8b", Namespace: "default", Status: "Running", Node: "minikube"},
			{Name: "redis-84d7fb", Namespace: "default", Status: "Running", Node: "minikube"},
		},
		Services: []FakeService{
			{Name: "nginx-service", Namespace: "default", Type: "ClusterIP", Port: 80},
			{Name: "postgres-service", Namespace: "default", Type: "ClusterIP", Port: 5432},
			{Name: "redis-service", Namespace: "default", Type: "ClusterIP", Port: 6379},
		},
		Deployments: []FakeDeployment{
			{Name: "nginx-deployment", Namespace: "default", Replicas: 2, Image: "nginx:1.21"},
			{Name: "postgres", Namespace: "default", Replicas: 1, Image: "postgres:13"},
			{Name: "redis", Namespace: "default", Replicas: 1, Image: "redis:6"},
		},
	}

	// Add a production-like cluster
	adapter.clusters["prod-cluster"] = &FakeCluster{
		Name:  "prod-cluster",
		Nodes: []string{"node-1", "node-2", "node-3"},
		Pods: []FakePod{
			{Name: "frontend-web-abc123", Namespace: "production", Status: "Running", Node: "node-1"},
			{Name: "frontend-web-def456", Namespace: "production", Status: "Running", Node: "node-2"},
			{Name: "api-backend-ghi789", Namespace: "production", Status: "Running", Node: "node-3"},
			{Name: "database-primary", Namespace: "production", Status: "Running", Node: "node-1"},
		},
		Services: []FakeService{
			{Name: "frontend-service", Namespace: "production", Type: "LoadBalancer", Port: 80},
			{Name: "api-service", Namespace: "production", Type: "ClusterIP", Port: 8080},
			{Name: "database-service", Namespace: "production", Type: "ClusterIP", Port: 5432},
		},
		Deployments: []FakeDeployment{
			{Name: "frontend-web", Namespace: "production", Replicas: 2, Image: "myapp/frontend:v1.2.3"},
			{Name: "api-backend", Namespace: "production", Replicas: 1, Image: "myapp/api:v1.2.3"},
			{Name: "database", Namespace: "production", Replicas: 1, Image: "postgres:14"},
		},
	}

	return adapter
}

func (a *FakeKubernetesAdapter) TestConnection(ctx context.Context, conn *domain.Connection) error {
	clusterName := conn.Subtype
	if clusterName == "" {
		clusterName = conn.Config.Extra["cluster"]
	}

	if _, exists := a.clusters[clusterName]; !exists {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	return nil
}

func (a *FakeKubernetesAdapter) FetchContext(ctx context.Context, conn *domain.Connection) (string, error) {
	clusterName := conn.Subtype
	if clusterName == "" {
		clusterName = conn.Config.Extra["cluster"]
	}

	cluster, exists := a.clusters[clusterName]
	if !exists {
		return "", fmt.Errorf("cluster '%s' not found", clusterName)
	}

	var context strings.Builder

	// Add cluster info
	context.WriteString(fmt.Sprintf("Cluster: %s\n", cluster.Name))
	context.WriteString(fmt.Sprintf("Nodes: %s\n", strings.Join(cluster.Nodes, ", ")))

	// Add pods
	context.WriteString("\nPods:\n")
	for _, pod := range cluster.Pods {
		context.WriteString(fmt.Sprintf("pod: %s (%s, %s)\n", pod.Name, pod.Namespace, pod.Status))
	}

	// Add services
	context.WriteString("\nServices:\n")
	for _, svc := range cluster.Services {
		context.WriteString(fmt.Sprintf("service: %s (%s, %s:%d)\n", svc.Name, svc.Namespace, svc.Type, svc.Port))
	}

	// Add deployments
	context.WriteString("\nDeployments:\n")
	for _, dep := range cluster.Deployments {
		context.WriteString(fmt.Sprintf("deployment: %s (%s, %d replicas, %s)\n",
			dep.Name, dep.Namespace, dep.Replicas, dep.Image))
	}

	return context.String(), nil
}

func (a *FakeKubernetesAdapter) ExecuteCommand(ctx context.Context, conn *domain.Connection, command string) (string, error) {
	clusterName := conn.Subtype
	if clusterName == "" {
		clusterName = conn.Config.Extra["cluster"]
	}

	cluster, exists := a.clusters[clusterName]
	if !exists {
		return "", fmt.Errorf("cluster '%s' not found", clusterName)
	}

	// Simulate kubectl commands
	if strings.Contains(command, "get pods") {
		var result strings.Builder
		result.WriteString("NAME                    READY   STATUS    RESTARTS   AGE\n")
		for _, pod := range cluster.Pods {
			result.WriteString(fmt.Sprintf("%-24s 1/1     %-8s 0          2d\n", pod.Name, pod.Status))
		}
		return result.String(), nil
	}

	if strings.Contains(command, "get services") {
		var result strings.Builder
		result.WriteString("NAME               TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE\n")
		for _, svc := range cluster.Services {
			result.WriteString(fmt.Sprintf("%-18s %-14s 10.96.%d.%d   <none>        %d/TCP   2d\n",
				svc.Name, svc.Type, (svc.Port%254)+1, (svc.Port%254)+100, svc.Port))
		}
		return result.String(), nil
	}

	if strings.Contains(command, "get deployments") {
		var result strings.Builder
		result.WriteString("NAME               READY   UP-TO-DATE   AVAILABLE   AGE\n")
		for _, dep := range cluster.Deployments {
			result.WriteString(fmt.Sprintf("%-18s %d/%d     %d            %d           2d\n",
				dep.Name, dep.Replicas, dep.Replicas, dep.Replicas, dep.Replicas))
		}
		return result.String(), nil
	}

	if strings.Contains(command, "get nodes") {
		var result strings.Builder
		result.WriteString("NAME       STATUS   ROLES           AGE   VERSION\n")
		for i, node := range cluster.Nodes {
			role := "worker"
			if i == 0 {
				role = "control-plane"
			}
			result.WriteString(fmt.Sprintf("%-10s Ready    %-15s 7d    v1.24.0\n", node, role))
		}
		return result.String(), nil
	}

	// Default response for unknown commands
	return fmt.Sprintf("Executed: %s\nResult: Command completed successfully on cluster %s", command, cluster.Name), nil
}

func (a *FakeKubernetesAdapter) GetCluster(name string) *FakeCluster {
	return a.clusters[name]
}

func (a *FakeKubernetesAdapter) AddCluster(cluster *FakeCluster) {
	a.clusters[cluster.Name] = cluster
}

func (a *FakeKubernetesAdapter) ListClusters() []string {
	var names []string
	for name := range a.clusters {
		names = append(names, name)
	}
	return names
}
