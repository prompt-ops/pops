package cloud

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

type AzureConnection struct {
	Connection      config.Connection
	ResourceGroups  []ResourceGroup
	VirtualMachines []VirtualMachine
	StorageAccounts []StorageAccount
}

func NewAzureConnection(conn config.Connection) *AzureConnection {
	return &AzureConnection{
		Connection: conn,
	}
}

// CheckAuthentication checks if the Azure connection is authenticated
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

// InitialContext sets up the initial context for the Azure connection
func (a *AzureConnection) InitialContext() error {
	// Get all resource groups
	resourceGroups, err := a.getResourceGroups()
	if err != nil {
		return err
	}
	a.ResourceGroups = resourceGroups

	// Get all VMs
	vms, err := a.getVirtualMachines()
	if err != nil {
		return err
	}
	a.VirtualMachines = vms

	// Get all storage accounts
	storageAccounts, err := a.getStorageAccounts()
	if err != nil {
		return err
	}
	a.StorageAccounts = storageAccounts

	return nil
}

// GetContext returns a string representation of the current state of the AzureConnection
func (a *AzureConnection) GetContext() string {
	var sb strings.Builder

	sb.WriteString("Azure Connection Context:\n")

	sb.WriteString("Resource Groups:\n")
	for _, rg := range a.ResourceGroups {
		sb.WriteString(fmt.Sprintf("- %s\n", rg.Name))
	}

	sb.WriteString("Virtual Machines:\n")
	for _, vm := range a.VirtualMachines {
		sb.WriteString(fmt.Sprintf("- %s (Resource Group: %s)\n", vm.Name, vm.ResourceGroup))
	}

	sb.WriteString("Storage Accounts:\n")
	for _, sa := range a.StorageAccounts {
		sb.WriteString(fmt.Sprintf("- %s (Resource Group: %s)\n", sa.Name, sa.ResourceGroup))
	}

	return sb.String()
}

// prettyPrintContext returns a nicely formatted string representation of the current state of the AzureConnection
func (a *AzureConnection) PrintContext() string {
	var buf bytes.Buffer

	// Print resource groups
	buf.WriteString("Resource Groups:\n")
	resourceGroupTable := tablewriter.NewWriter(&buf)
	resourceGroupTable.SetHeader([]string{"Name"})
	for _, rg := range a.ResourceGroups {
		resourceGroupTable.Append([]string{rg.Name})
	}
	resourceGroupTable.Render()

	// Print virtual machines
	buf.WriteString("\nVirtual Machines:\n")
	vmTable := tablewriter.NewWriter(&buf)
	vmTable.SetHeader([]string{"Name", "Resource Group"})
	for _, vm := range a.VirtualMachines {
		vmTable.Append([]string{vm.Name, vm.ResourceGroup})
	}
	vmTable.Render()

	// Print storage accounts
	buf.WriteString("\nStorage Accounts:\n")
	storageAccountTable := tablewriter.NewWriter(&buf)
	storageAccountTable.SetHeader([]string{"Name", "Resource Group"})
	for _, sa := range a.StorageAccounts {
		storageAccountTable.Append([]string{sa.Name, sa.ResourceGroup})
	}
	storageAccountTable.Render()

	return buf.String()
}

func (a *AzureConnection) GetCommand(prompt string) (string, error) {
	cmd, err := ai.GetCommand(prompt, a.CommandType(), "")
	if err != nil {
		return "", fmt.Errorf("failed to get command from AI: %v", err)
	}

	return cmd.Command, nil
}

func (a *AzureConnection) ExecuteCommand(command string) ([]byte, error) {
	cmd := exec.Command(command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	return output, nil
}

// ResourceGroup represents an Azure resource group
type ResourceGroup struct {
	Name string `json:"name"`
}

// VirtualMachine represents an Azure virtual machine
type VirtualMachine struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
}

// StorageAccount represents an Azure storage account
type StorageAccount struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
}

// getResourceGroups gets all resource groups
func (a *AzureConnection) getResourceGroups() ([]ResourceGroup, error) {
	cmd := exec.Command("az", "group", "list", "--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %v", string(output))
	}

	var resourceGroups []ResourceGroup
	if err := json.Unmarshal(output, &resourceGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource groups: %v", err)
	}

	return resourceGroups, nil
}

// getVirtualMachines retrieves all virtual machines without specifying a subscription
func (a *AzureConnection) getVirtualMachines() ([]VirtualMachine, error) {
	cmd := exec.Command("az", "vm", "list", "--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual machines: %v. Output: %s", err, string(output))
	}

	// Validate JSON
	if !json.Valid(output) {
		return nil, fmt.Errorf("invalid JSON received from az vm list: %s", string(output))
	}

	var vms []VirtualMachine
	if err := json.Unmarshal(output, &vms); err != nil {
		return nil, fmt.Errorf("failed to parse virtual machines: %v. Output: %s", err, string(output))
	}

	return vms, nil
}

// getStorageAccounts gets all storage accounts
func (a *AzureConnection) getStorageAccounts() ([]StorageAccount, error) {
	cmd := exec.Command("az", "storage", "account", "list", "--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list storage accounts: %v", string(output))
	}

	var storageAccounts []StorageAccount
	if err := json.Unmarshal(output, &storageAccounts); err != nil {
		return nil, fmt.Errorf("failed to parse storage accounts: %v", err)
	}

	return storageAccounts, nil
}

// Type returns the type of the connection
func (a *AzureConnection) Type() string {
	return "cloud"
}

// SubType returns the subtype of the connection
func (a *AzureConnection) SubType() string {
	return "azure"
}

// CommandType returns the command type for the connection
func (a *AzureConnection) CommandType() string {
	return "az cli command"
}
