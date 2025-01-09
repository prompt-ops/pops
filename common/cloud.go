package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/prompt-ops/pops/ai"
)

var (
	// TODO: Rename?

	// AzureCloudConnection
	AzureCloudConnection = AvailableCloudConnectionType{
		Subtype: "Azure",
	}

	// AWSCloudConnection
	AWSCloudConnection = AvailableCloudConnectionType{
		Subtype: "AWS",
	}

	// GCPCloudConnection
	GCPCloudConnection = AvailableCloudConnectionType{
		Subtype: "GCP",
	}

	// AvailableCloudConnectionTypes is a list of available cloud connections.
	AvailableCloudConnectionTypes = []AvailableCloudConnectionType{
		AzureCloudConnection,
		AWSCloudConnection,
		GCPCloudConnection,
	}
)

// AvailableCloudConnectionType is a helper struct to UI to list available cloud connection types.
// Subtype will be shown in the UI.
type AvailableCloudConnectionType struct {
	Subtype string
}

type CloudConnectionType struct {
	// MainType of the connection type.
	// Example: "cloud".
	MainType string `json:"mainType"`

	// Subtype of the cloud connection type.
	// Example: "aws", "gcp", "azure".
	Subtype string `json:"subtype"`
}

func (c CloudConnectionType) GetMainType() string {
	return "Cloud"
}

func (c CloudConnectionType) GetSubtype() string {
	return c.Subtype
}

type CloudConnectionDetails struct {
}

func (c CloudConnectionDetails) GetDriver() string {
	return ""
}

// NewCloudConnection creates a new cloud connection.
func NewCloudConnection(name string, cloudProvider AvailableCloudConnectionType) Connection {
	return Connection{
		Name: name,
		Type: CloudConnectionType{
			MainType: "Cloud",
			Subtype:  cloudProvider.Subtype,
		},
		Details: CloudConnectionDetails{},
	}
}

// GetCloudConnectionDetails retrieves the CloudConnectionDetails from a Connection.
func GetCloudConnectionDetails(conn Connection) (CloudConnectionDetails, error) {
	if conn.Type.GetMainType() != ConnectionTypeCloud {
		return CloudConnectionDetails{}, fmt.Errorf("connection is not of type 'cloud'")
	}
	details, ok := conn.Details.(CloudConnectionDetails)
	if !ok {
		return CloudConnectionDetails{}, fmt.Errorf("invalid connection details for 'cloud'")
	}
	return details, nil
}

// BaseCloudConnection is a partial implementation of the ConnectionInterface for cloud.
type BaseCloudConnection struct {
	Connection Connection
}

func (c *BaseCloudConnection) GetConnection() Connection {
	return c.Connection
}

func (c *BaseCloudConnection) ExecuteCommand(command string) ([]byte, error) {
	// Split the command into command and arguments
	// This is required for exec.Command
	// Example: "az group list" -> "az", "group", "list"
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

func (c *BaseCloudConnection) FormatResultAsTable(result []byte) (string, error) {
	// Parse the JSON data
	var rows []map[string]interface{}
	if err := json.Unmarshal(result, &rows); err != nil {
		return "", fmt.Errorf("failed to parse JSON result: %v", err)
	}

	// Check if there is any data to format
	if len(rows) == 0 {
		return "No data available", nil
	}

	// Extract the header from the first row
	var header []string
	for col := range rows[0] {
		header = append(header, col)
	}

	// Prepare rows for the table
	var tableRows [][]string
	for _, row := range rows {
		var tableRow []string
		for _, col := range header {
			// Handle nil values or values of different types
			if value, ok := row[col]; ok {
				tableRow = append(tableRow, fmt.Sprintf("%v", value))
			} else {
				tableRow = append(tableRow, "") // Empty for missing values
			}
		}
		tableRows = append(tableRows, tableRow)
	}

	// Create a buffer to write the formatted table
	var buffer bytes.Buffer
	table := tablewriter.NewWriter(&buffer)

	// Add the header and rows to the table
	table.SetHeader(header)
	for _, row := range tableRows {
		table.Append(row)
	}

	// Render the table
	table.Render()

	return buffer.String(), nil
}

type AzureConnection struct {
	BaseCloudConnection

	ResourceGroups []AzureResourceGroup
}

func (a *AzureConnection) CheckAuthentication() error {
	// Check if az cli is installed
	if _, err := exec.LookPath("az"); err != nil {
		return fmt.Errorf("az CLI is not installed")
	}

	// Check if az cli is logged in
	cmd := exec.Command("az", "account", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("az CLI is not logged in: %v", string(output))
	}

	return nil
}

// SetContext sets the context for the Azure connection.
// This will populate the resource groups.
func (a *AzureConnection) SetContext() error {
	// Get all resource groups
	resourceGroups, err := a.getResourceGroups()
	if err != nil {
		return err
	}

	a.ResourceGroups = resourceGroups
	return nil
}

// GetContext returns the resource groups in the Azure connection.
func (a *AzureConnection) GetContext() string {
	if a.ResourceGroups == nil {
		// Call SetContext to populate the resource groups.
		// This is a fallback in case SetContext is not called.
		if err := a.SetContext(); err != nil {
			return fmt.Sprintf("Error getting context: %v", err)
		}
	}

	context := fmt.Sprintf("%s Connection Details:\n", a.Connection.Type.GetSubtype())
	context += "Resource Groups:\n"

	for _, rg := range a.ResourceGroups {
		context += fmt.Sprintf("- %s\n", rg.Name)
	}

	return context
}

func NewAzureConnection(connnection *Connection) *AzureConnection {
	return &AzureConnection{
		BaseCloudConnection: BaseCloudConnection{
			Connection: *connnection,
		},
	}
}

func (a *AzureConnection) GetCommand(prompt string) (string, error) {
	if a.ResourceGroups == nil {
		// Call GetContext to populate the resource groups.
		// This is a fallback in case GetContext is not called.
		if err := a.SetContext(); err != nil {
			return "", fmt.Errorf("error getting context: %v", err)
		}
	}

	// Because this is the initial version of Prompt-Ops,
	// we are going to have overlaps like having context both
	// in the connection and in the AI model.
	// As we iterate on building Prompt-Ops, we will remove this overlap.
	aiModel, err := ai.NewOpenAIModel(a.CommandType(), a.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	cmd, err := aiModel.GetCommand(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get command from AI: %v", err)
	}

	return cmd.Command, nil
}

func (a *AzureConnection) GetAnswer(prompt string) (string, error) {
	if a.ResourceGroups == nil {
		// Call GetContext to populate the resource groups.
		// This is a fallback in case GetContext is not called.
		if err := a.SetContext(); err != nil {
			return "", fmt.Errorf("error getting context: %v", err)
		}
	}

	// Because this is the initial version of Prompt-Ops,
	// we are going to have overlaps like having context both
	// in the connection and in the AI model.
	// As we iterate on building Prompt-Ops, we will remove this overlap.
	aiModel, err := ai.NewOpenAIModel(a.CommandType(), a.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	answer, err := aiModel.GetAnswer(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get answer from AI: %v", err)
	}

	return answer.Answer, nil
}

func (a *AzureConnection) CommandType() string {
	return "az cli command"
}

// AzureResourceGroup represents an Azure resource group.
type AzureResourceGroup struct {
	Name string `json:"name"`
}

// getResourceGroups gets all Azure resource groups.
func (a *AzureConnection) getResourceGroups() ([]AzureResourceGroup, error) {
	cmd := exec.Command("az", "group", "list", "--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %v", string(output))
	}

	var resourceGroups []AzureResourceGroup
	if err := json.Unmarshal(output, &resourceGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource groups: %v", err)
	}

	return resourceGroups, nil
}
