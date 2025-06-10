package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/internal/domain"
)

type FakeCloudAdapter struct {
	providers map[string]*FakeCloudProvider
}

type FakeCloudProvider struct {
	Name      string
	VMs       []FakeVM
	Storage   []FakeStorage
	Networks  []FakeNetwork
}

type FakeVM struct {
	Name   string
	Status string
	Size   string
	Region string
}

type FakeStorage struct {
	Name string
	Size string
	Type string
}

type FakeNetwork struct {
	Name   string
	CIDR   string
	Region string
}

func NewFakeCloudAdapter() *FakeCloudAdapter {
	adapter := &FakeCloudAdapter{
		providers: make(map[string]*FakeCloudProvider),
	}

	// Create fake cloud resources
	adapter.providers["aws"] = &FakeCloudProvider{
		Name: "aws",
		VMs: []FakeVM{
			{Name: "web-server-1", Status: "running", Size: "t3.medium", Region: "us-east-1"},
			{Name: "api-server-1", Status: "running", Size: "t3.large", Region: "us-east-1"},
			{Name: "worker-1", Status: "stopped", Size: "t3.small", Region: "us-west-2"},
		},
		Storage: []FakeStorage{
			{Name: "app-data", Size: "100GB", Type: "gp3"},
			{Name: "backup-storage", Size: "500GB", Type: "s3"},
		},
		Networks: []FakeNetwork{
			{Name: "vpc-prod", CIDR: "10.0.0.0/16", Region: "us-east-1"},
			{Name: "vpc-dev", CIDR: "10.1.0.0/16", Region: "us-west-2"},
		},
	}

	return adapter
}

func (a *FakeCloudAdapter) TestConnection(ctx context.Context, conn *domain.Connection) error {
	return nil
}

func (a *FakeCloudAdapter) FetchContext(ctx context.Context, conn *domain.Connection) (string, error) {
	return a.GetContext(ctx, conn)
}

func (a *FakeCloudAdapter) ExecuteCommand(ctx context.Context, conn *domain.Connection, command string) (string, error) {
	provider, exists := a.providers[conn.Subtype]
	if !exists {
		return "", fmt.Errorf("cloud provider %s not found", conn.Subtype)
	}

	command = strings.ToLower(strings.TrimSpace(command))

	if strings.Contains(command, "ec2") && strings.Contains(command, "describe-instances") {
		return a.listVMs(provider), nil
	}
	if strings.Contains(command, "s3") && strings.Contains(command, "ls") {
		return a.listStorage(provider), nil
	}
	if strings.Contains(command, "vpc") && strings.Contains(command, "describe") {
		return a.listNetworks(provider), nil
	}

	return "Command executed successfully (mock)", nil
}

func (a *FakeCloudAdapter) GetContext(ctx context.Context, conn *domain.Connection) (string, error) {
	provider, exists := a.providers[conn.Subtype]
	if !exists {
		return "", fmt.Errorf("cloud provider %s not found", conn.Subtype)
	}

	var context strings.Builder
	context.WriteString(fmt.Sprintf("Cloud Provider: %s\n", provider.Name))
	
	context.WriteString("Virtual Machines:\n")
	for _, vm := range provider.VMs {
		context.WriteString(fmt.Sprintf("vm: %s\n", vm.Name))
	}
	
	context.WriteString("Storage:\n")
	for _, storage := range provider.Storage {
		context.WriteString(fmt.Sprintf("storage: %s\n", storage.Name))
	}
	
	context.WriteString("Networks:\n")
	for _, network := range provider.Networks {
		context.WriteString(fmt.Sprintf("network: %s\n", network.Name))
	}

	return context.String(), nil
}

func (a *FakeCloudAdapter) listVMs(provider *FakeCloudProvider) string {
	var result strings.Builder
	result.WriteString("INSTANCE ID | NAME | STATUS | SIZE | REGION\n")
	for i, vm := range provider.VMs {
		result.WriteString(fmt.Sprintf("i-%d | %s | %s | %s | %s\n", 
			1000+i, vm.Name, vm.Status, vm.Size, vm.Region))
	}
	return result.String()
}

func (a *FakeCloudAdapter) listStorage(provider *FakeCloudProvider) string {
	var result strings.Builder
	result.WriteString("BUCKET/VOLUME | SIZE | TYPE\n")
	for _, storage := range provider.Storage {
		result.WriteString(fmt.Sprintf("%s | %s | %s\n", 
			storage.Name, storage.Size, storage.Type))
	}
	return result.String()
}

func (a *FakeCloudAdapter) listNetworks(provider *FakeCloudProvider) string {
	var result strings.Builder
	result.WriteString("VPC ID | NAME | CIDR | REGION\n")
	for i, network := range provider.Networks {
		result.WriteString(fmt.Sprintf("vpc-%d | %s | %s | %s\n", 
			100+i, network.Name, network.CIDR, network.Region))
	}
	return result.String()
}