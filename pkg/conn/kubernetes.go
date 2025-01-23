package conn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/prompt-ops/pops/pkg/ai"
)

var (
	// AvailableDatabaseConnectionTypes is a list of available database connections.
	AvailableKubernetesConnectionTypes = []AvailableKubernetesConnectionType{}
)

// AvailableKubernetesConnection is a helper struct to UI to list available kubernetes connection types.
// Subtype will be shown in the UI.
type AvailableKubernetesConnectionType struct {
	Subtype string
}

type KubernetesConnectionType struct {
	// MainType of the connection type.
	// Example: "kubernetes".
	MainType string `json:"mainType"`
}

func (k KubernetesConnectionType) GetMainType() string {
	return ConnectionTypeKubernetes
}

func (k KubernetesConnectionType) GetSubtype() string {
	return ""
}

type KubernetesConnectionDetails struct {
	// SelectedContext is the selected context for the kubernetes connection.
	SelectedContext string `json:"selectedContext"`
}

func (k KubernetesConnectionDetails) GetDriver() string {
	return ""
}

func (k KubernetesConnectionDetails) GetSelectedContext() string {
	return k.SelectedContext
}

// NewKubernetesConnection creates a new Kubernetes connection.
func NewKubernetesConnection(name, context string) Connection {
	return Connection{
		Name: name,
		Type: KubernetesConnectionType{
			MainType: ConnectionTypeKubernetes,
		},
		Details: KubernetesConnectionDetails{
			SelectedContext: context,
		},
	}
}

// GetKubernetesConnectionDetails retrieves the KubernetesConnectionDetails from a Connection.
func GetKubernetesConnectionDetails(conn Connection) (KubernetesConnectionDetails, error) {
	if conn.Type.GetMainType() != ConnectionTypeKubernetes {
		return KubernetesConnectionDetails{}, fmt.Errorf("connection is not of type 'kubernetes'")
	}
	details, ok := conn.Details.(KubernetesConnectionDetails)
	if !ok {
		return KubernetesConnectionDetails{}, fmt.Errorf("invalid connection details for 'kubernetes'")
	}
	return details, nil
}

// KubernetesConnection is the implementation of the ConnectionInterface for Kubernetes.
type KubernetesConnectionImpl struct {
	Connection Connection

	Namespaces  []Namespace
	Pods        []Pod
	Deployments []Deployment
	Services    []Service
}

func NewKubernetesConnectionImpl(connection *Connection) *KubernetesConnectionImpl {
	return &KubernetesConnectionImpl{
		Connection: *connection,
	}
}

var _ ConnectionInterface = &KubernetesConnectionImpl{}

func (k *KubernetesConnectionImpl) GetConnection() Connection {
	return k.Connection
}

func (k *KubernetesConnectionImpl) CheckAuthentication() error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl is not installed")
	}

	cmd := exec.Command("kubectl", "cluster-info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to access Kubernetes cluster: %v. Output: %s", err, string(output))
	}

	return nil
}

func (k *KubernetesConnectionImpl) SetContext() error {
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

func (k *KubernetesConnectionImpl) GetContext() string {
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

func (k *KubernetesConnectionImpl) GetFormattedContext() (string, error) {
	var buffer bytes.Buffer
	table := tablewriter.NewWriter(&buffer)

	// Namespaces
	table.SetHeader([]string{"Namespaces"})
	for _, ns := range k.Namespaces {
		table.Append([]string{ns.Name})
	}
	table.Render()

	// Pods
	table = tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{"Pods", "Namespace"})
	for _, pod := range k.Pods {
		table.Append([]string{pod.Name, pod.Namespace})
	}
	table.Render()

	// Deployments
	table = tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{"Deployments", "Namespace"})
	for _, dep := range k.Deployments {
		table.Append([]string{dep.Name, dep.Namespace})
	}
	table.Render()

	// Services
	table = tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{"Services", "Namespace"})
	for _, svc := range k.Services {
		table.Append([]string{svc.Name, svc.Namespace})
	}
	table.Render()

	return buffer.String(), nil
}

func (k *KubernetesConnectionImpl) GetCommand(prompt string) (string, error) {
	aiModel, err := ai.NewOpenAIModel(k.CommandType(), k.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	cmd, err := aiModel.GetCommand(prompt)
	if err != nil {
		return "", err
	}

	return cmd.Command, nil
}

func (k *KubernetesConnectionImpl) GetAnswer(prompt string) (string, error) {
	aiModel, err := ai.NewOpenAIModel(k.CommandType(), k.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	answer, err := aiModel.GetAnswer(prompt)
	if err != nil {
		return "", err
	}

	return answer.Answer, nil
}

func (k *KubernetesConnectionImpl) ExecuteCommand(command string) ([]byte, error) {
	// Split the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	// The first part is the command, the rest are the arguments
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	return output, nil
}

func (k *KubernetesConnectionImpl) FormatResultAsTable(result []byte) (string, error) {
	resultStr := string(result)

	lines := strings.Split(resultStr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no data to format")
	}

	header := strings.Fields(lines[0])
	if len(header) == 0 {
		return "", fmt.Errorf("no headers found in result")
	}

	var rows [][]string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		row := strings.Fields(line)
		rows = append(rows, row)
	}

	var buffer bytes.Buffer

	table := tablewriter.NewWriter(&buffer)
	table.SetAutoWrapText(false)         // Disable wrapping
	table.SetReflowDuringAutoWrap(false) // Avoid reflows
	table.SetRowLine(true)               // Horizontal line between rows
	table.SetAutoFormatHeaders(false)    // Keep header text as-is
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(header)

	for _, row := range rows {
		table.Append(row)
	}

	table.Render()

	return buffer.String(), nil
}

func (k *KubernetesConnectionImpl) CommandType() string {
	return "kubectl command"
}

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
func (k *KubernetesConnectionImpl) getNamespaces() ([]Namespace, error) {
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
func (k *KubernetesConnectionImpl) getPods() ([]Pod, error) {
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
func (k *KubernetesConnectionImpl) getDeployments() ([]Deployment, error) {
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
func (k *KubernetesConnectionImpl) getServices() ([]Service, error) {
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
