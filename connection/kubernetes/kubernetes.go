package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/prompt-ops/pops/ai"
	config "github.com/prompt-ops/pops/config"
)

// KubernetesConnection manages the connection and context to a Kubernetes cluster
type KubernetesConnection struct {
	Connection  config.Connection
	Namespaces  []Namespace
	Pods        []Pod
	Deployments []Deployment
	Services    []Service
}

// NewKubernetesConnection initializes a new KubernetesConnection
func NewKubernetesConnection(conn config.Connection) *KubernetesConnection {
	return &KubernetesConnection{
		Connection: conn,
	}
}

// CheckAuthentication verifies if kubectl is installed and configured correctly
func (k *KubernetesConnection) CheckAuthentication() error {
	// Check if kubectl is installed
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl is not installed")
	}

	// Check if kubectl can access the cluster
	cmd := exec.Command("kubectl", "cluster-info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to access Kubernetes cluster: %v. Output: %s", err, string(output))
	}

	return nil
}

// InitialContext sets up the initial context by fetching namespaces, pods, deployments, and services
func (k *KubernetesConnection) InitialContext() error {
	// Get all namespaces
	namespaces, err := k.getNamespaces()
	if err != nil {
		return err
	}
	k.Namespaces = namespaces

	// Get all pods
	pods, err := k.getPods()
	if err != nil {
		return err
	}
	k.Pods = pods

	// Get all deployments
	deployments, err := k.getDeployments()
	if err != nil {
		return err
	}
	k.Deployments = deployments

	// Get all services
	services, err := k.getServices()
	if err != nil {
		return err
	}
	k.Services = services

	return nil
}

// GetContext returns a string representation of the current state of the KubernetesConnection
func (k *KubernetesConnection) GetContext() string {
	var sb strings.Builder

	sb.WriteString("Kubernetes Connection Context:\n\n")

	sb.WriteString("Namespaces:\n")
	for _, ns := range k.Namespaces {
		sb.WriteString(fmt.Sprintf("- %s\n", ns.Name))
	}

	sb.WriteString("\nPods:\n")
	for _, pod := range k.Pods {
		sb.WriteString(fmt.Sprintf("- %s (Namespace: %s)\n", pod.Name, pod.Namespace))
	}

	sb.WriteString("\nDeployments:\n")
	for _, dep := range k.Deployments {
		sb.WriteString(fmt.Sprintf("- %s (Namespace: %s)\n", dep.Name, dep.Namespace))
	}

	sb.WriteString("\nServices:\n")
	for _, svc := range k.Services {
		sb.WriteString(fmt.Sprintf("- %s (Namespace: %s)\n", svc.Name, svc.Namespace))
	}

	return sb.String()
}

// PrintContext returns a nicely formatted string representation of the current state of the KubernetesConnection
func (k *KubernetesConnection) PrintContext() string {
	var buf bytes.Buffer

	// Print namespaces
	buf.WriteString("Namespaces:\n")
	namespaceTable := tablewriter.NewWriter(&buf)
	namespaceTable.SetHeader([]string{"Name"})
	for _, ns := range k.Namespaces {
		namespaceTable.Append([]string{ns.Name})
	}
	namespaceTable.Render()

	// Print pods
	buf.WriteString("\nPods:\n")
	podTable := tablewriter.NewWriter(&buf)
	podTable.SetHeader([]string{"Name", "Namespace"})
	for _, pod := range k.Pods {
		podTable.Append([]string{pod.Name, pod.Namespace})
	}
	podTable.Render()

	// Print deployments
	buf.WriteString("\nDeployments:\n")
	deploymentTable := tablewriter.NewWriter(&buf)
	deploymentTable.SetHeader([]string{"Name", "Namespace"})
	for _, dep := range k.Deployments {
		deploymentTable.Append([]string{dep.Name, dep.Namespace})
	}
	deploymentTable.Render()

	// Print services
	buf.WriteString("\nServices:\n")
	serviceTable := tablewriter.NewWriter(&buf)
	serviceTable.SetHeader([]string{"Name", "Namespace"})
	for _, svc := range k.Services {
		serviceTable.Append([]string{svc.Name, svc.Namespace})
	}
	serviceTable.Render()

	return buf.String()
}

func (k *KubernetesConnection) GetCommand(prompt string) (string, error) {
	cmd, err := ai.GetCommand(prompt, k.CommandType(), k.GetContext())
	if err != nil {
		return "", err
	}

	return cmd.Command, nil
}

func (k *KubernetesConnection) ExecuteCommand(command string) ([]byte, error) {
	cmd := exec.Command(command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	return output, nil
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name string `json:"name"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Deployment represents a Kubernetes deployment
type Deployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Service represents a Kubernetes service
type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// getNamespaces retrieves all namespaces in the cluster
func (k *KubernetesConnection) getNamespaces() ([]Namespace, error) {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v. Output: %s", err, string(output))
	}

	var nsList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &nsList); err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %v. Output: %s", err, string(output))
	}

	var namespaces []Namespace
	for _, item := range nsList.Items {
		namespaces = append(namespaces, Namespace{Name: item.Metadata.Name})
	}

	return namespaces, nil
}

// getPods retrieves all pods across all namespaces
func (k *KubernetesConnection) getPods() ([]Pod, error) {
	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v. Output: %s", err, string(output))
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %v. Output: %s", err, string(output))
	}

	var pods []Pod
	for _, item := range podList.Items {
		pods = append(pods, Pod{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
		})
	}

	return pods, nil
}

// getDeployments retrieves all deployments across all namespaces
func (k *KubernetesConnection) getDeployments() ([]Deployment, error) {
	cmd := exec.Command("kubectl", "get", "deployments", "--all-namespaces", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v. Output: %s", err, string(output))
	}

	var depList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &depList); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %v. Output: %s", err, string(output))
	}

	var deployments []Deployment
	for _, item := range depList.Items {
		deployments = append(deployments, Deployment{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
		})
	}

	return deployments, nil
}

// getServices retrieves all services across all namespaces
func (k *KubernetesConnection) getServices() ([]Service, error) {
	cmd := exec.Command("kubectl", "get", "services", "--all-namespaces", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v. Output: %s", err, string(output))
	}

	var svcList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &svcList); err != nil {
		return nil, fmt.Errorf("failed to parse services: %v. Output: %s", err, string(output))
	}

	var services []Service
	for _, item := range svcList.Items {
		services = append(services, Service{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
		})
	}

	return services, nil
}

// Type returns the type of the connection
func (k *KubernetesConnection) Type() string {
	return "cloud"
}

// SubType returns the subtype of the connection
func (k *KubernetesConnection) SubType() string {
	return "kubernetes"
}

// CommandType returns the command type for the connection
func (k *KubernetesConnection) CommandType() string {
	return "kubernetes-command"
}
