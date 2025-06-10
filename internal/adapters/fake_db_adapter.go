package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/prompt-ops/pops/internal/domain"
)

type FakeDatabaseAdapter struct {
	databases map[string]*FakeDatabase
}

type FakeDatabase struct {
	Name   string
	Tables []FakeTable
}

type FakeTable struct {
	Name    string
	Columns []FakeColumn
}

type FakeColumn struct {
	Name string
	Type string
}

func NewFakeDatabaseAdapter() *FakeDatabaseAdapter {
	adapter := &FakeDatabaseAdapter{
		databases: make(map[string]*FakeDatabase),
	}

	// Create fake databases
	adapter.databases["postgres"] = &FakeDatabase{
		Name: "postgres",
		Tables: []FakeTable{
			{
				Name: "users",
				Columns: []FakeColumn{
					{Name: "id", Type: "bigint"},
					{Name: "email", Type: "varchar(255)"},
					{Name: "created_at", Type: "timestamp"},
				},
			},
			{
				Name: "orders",
				Columns: []FakeColumn{
					{Name: "id", Type: "bigint"},
					{Name: "user_id", Type: "bigint"},
					{Name: "total", Type: "decimal(10,2)"},
					{Name: "status", Type: "varchar(50)"},
				},
			},
		},
	}

	return adapter
}

func (a *FakeDatabaseAdapter) TestConnection(ctx context.Context, conn *domain.Connection) error {
	return nil
}

func (a *FakeDatabaseAdapter) FetchContext(ctx context.Context, conn *domain.Connection) (string, error) {
	return a.GetContext(ctx, conn)
}

func (a *FakeDatabaseAdapter) ExecuteCommand(ctx context.Context, conn *domain.Connection, command string) (string, error) {
	database, exists := a.databases[conn.Subtype]
	if !exists {
		return "", fmt.Errorf("database %s not found", conn.Subtype)
	}

	command = strings.ToLower(strings.TrimSpace(command))

	if strings.HasPrefix(command, "select") {
		return a.executeSelect(command, database)
	}
	if strings.HasPrefix(command, "\\dt") || strings.Contains(command, "show tables") {
		return a.listTables(database), nil
	}
	if strings.HasPrefix(command, "\\d ") || strings.Contains(command, "describe") {
		tableName := extractTableName(command)
		return a.describeTable(database, tableName)
	}

	return "Query executed successfully (mock)", nil
}

func (a *FakeDatabaseAdapter) GetContext(ctx context.Context, conn *domain.Connection) (string, error) {
	database, exists := a.databases[conn.Subtype]
	if !exists {
		return "", fmt.Errorf("database %s not found", conn.Subtype)
	}

	var context strings.Builder
	context.WriteString(fmt.Sprintf("Database: %s\n", database.Name))
	
	for _, table := range database.Tables {
		context.WriteString(fmt.Sprintf("Table: %s\n", table.Name))
		for _, column := range table.Columns {
			context.WriteString(fmt.Sprintf("  - %s (%s)\n", column.Name, column.Type))
		}
	}

	return context.String(), nil
}

func (a *FakeDatabaseAdapter) executeSelect(command string, database *FakeDatabase) (string, error) {
	if strings.Contains(command, "users") {
		return "id | email | created_at\n1 | user@example.com | 2023-01-01 12:00:00\n2 | admin@example.com | 2023-01-02 10:00:00", nil
	}
	if strings.Contains(command, "orders") {
		return "id | user_id | total | status\n1 | 1 | 99.99 | completed\n2 | 2 | 149.50 | pending", nil
	}
	return "No results found", nil
}

func (a *FakeDatabaseAdapter) listTables(database *FakeDatabase) string {
	var result strings.Builder
	result.WriteString("Tables:\n")
	for _, table := range database.Tables {
		result.WriteString(fmt.Sprintf("- %s\n", table.Name))
	}
	return result.String()
}

func (a *FakeDatabaseAdapter) describeTable(database *FakeDatabase, tableName string) (string, error) {
	for _, table := range database.Tables {
		if table.Name == tableName {
			var result strings.Builder
			result.WriteString(fmt.Sprintf("Table: %s\n", table.Name))
			result.WriteString("Columns:\n")
			for _, column := range table.Columns {
				result.WriteString(fmt.Sprintf("- %s (%s)\n", column.Name, column.Type))
			}
			return result.String(), nil
		}
	}
	return "", fmt.Errorf("table %s not found", tableName)
}

func extractTableName(command string) string {
	parts := strings.Fields(command)
	for i, part := range parts {
		if part == "\\d" && i+1 < len(parts) {
			return parts[i+1]
		}
		if strings.Contains(part, "describe") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}