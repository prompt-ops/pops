package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
	"github.com/prompt-ops/pops/ai"
	"github.com/prompt-ops/pops/config"
)

type PostgreSQLConnection struct {
	Connection        *config.Connection
	ConnectionDetails *config.DatabaseConnectionDetails
	DB                *sql.DB
	TablesAndColumns  map[string][]ColumnDetail
}

type ColumnDetail struct {
	Name     string
	DataType string
}

func NewPostgreSQLConnection(conn *config.Connection) *PostgreSQLConnection {
	fmt.Println("Creating new PostgreSQL connection")
	databaseConnectionDetails, err := config.GetDatabaseConnectionDetails(*conn)
	if err != nil {
		panic(err)
	}

	return &PostgreSQLConnection{
		Connection:        conn,
		ConnectionDetails: &databaseConnectionDetails,
		TablesAndColumns:  make(map[string][]ColumnDetail),
	}
}

func (c *PostgreSQLConnection) Connect() error {
	db, err := sql.Open("postgres", c.ConnectionDetails.ConnectionString)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging the database: %v", err)
	}

	c.DB = db
	return nil
}

func (c *PostgreSQLConnection) Disconnect() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

func (c *PostgreSQLConnection) CheckAuthentication() error {
	db, err := sql.Open("postgres", c.ConnectionDetails.ConnectionString)
	if err != nil {
		return fmt.Errorf("Error connecting to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Error pinging the database: %v", err)
	}

	c.DB = db
	return nil
}

func (c *PostgreSQLConnection) InitialContext() error {
	if c.DB == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	query := `
        SELECT table_schema, table_name, column_name, data_type
        FROM information_schema.columns
        WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
        ORDER BY table_schema, table_name, ordinal_position;
    `

	rows, err := c.DB.Query(query)
	if err != nil {
		return fmt.Errorf("error querying database schema: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, table, column, dataType string
		if err := rows.Scan(&schema, &table, &column, &dataType); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		// check if the column is camel case
		// if so, put it in between quotes
		if hasAnyUpperCaseLetter(column) {
			column = fmt.Sprintf("\"%s\"", column)
		}

		fullTableName := fmt.Sprintf("%s.%s", schema, table)
		c.TablesAndColumns[fullTableName] = append(c.TablesAndColumns[fullTableName], ColumnDetail{
			Name:     column,
			DataType: dataType,
		})
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %v", err)
	}

	return nil
}

func hasAnyUpperCaseLetter(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func (c *PostgreSQLConnection) MaskedConnectionString() string {
	parsedURL, err := url.Parse(c.ConnectionDetails.ConnectionString)
	if err != nil {
		// If parsing fails, return the original connection string
		return c.ConnectionDetails.ConnectionString
	}

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		if _, hasPassword := parsedURL.User.Password(); hasPassword {
			parsedURL.User = url.UserPassword(username, "****")
		}
	}

	return parsedURL.String()
}

func (c *PostgreSQLConnection) GetContext() string {
	// Start with connection details
	context := fmt.Sprintf("PostgreSQL Connection Details:\n")
	context += fmt.Sprintf("Connection String: %s\n\n", c.MaskedConnectionString())

	// Add database schema information
	context += "Database Schema:\n"

	// Check if TablesAndColumns is populated
	if len(c.TablesAndColumns) == 0 {
		context += "No tables found or InitialContext() not called.\n"
		return context
	}

	// Iterate over each table and its columns
	for table, columns := range c.TablesAndColumns {
		context += fmt.Sprintf("- **%s**:\n", table)
		for _, column := range columns {
			context += fmt.Sprintf("  - `%s` (%s)\n", column.Name, column.DataType)
		}
	}

	return context
}

func (c *PostgreSQLConnection) PrintContext() string {
	// Start with connection details
	fmt.Println("PostgreSQL Connection Details:")
	fmt.Printf("Connection String: `%s`\n\n", c.MaskedConnectionString())

	// Add database schema information
	fmt.Println("Database Schema:")

	// Check if TablesAndColumns is populated
	if len(c.TablesAndColumns) == 0 {
		fmt.Println("No tables found or InitialContext() not called.")
		return ""
	}

	// Iterate over each table and its columns
	for table, columns := range c.TablesAndColumns {
		fmt.Printf("\nTable: %s\n", table)
		data := [][]string{}
		for _, column := range columns {
			data = append(data, []string{column.Name, column.DataType})
		}

		tableInstance := tablewriter.NewWriter(os.Stdout)
		tableInstance.SetHeader([]string{"Column Name", "Data Type"})
		tableInstance.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		tableInstance.SetCenterSeparator("|")
		tableInstance.AppendBulk(data)
		tableInstance.Render()
	}

	return ""
}

func (c *PostgreSQLConnection) GetCommand(prompt string) (string, error) {
	cmd, err := ai.GetCommand(prompt, c.CommandType(), c.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to get command from AI: %v", err)
	}

	return cmd.Command, nil
}

func (c *PostgreSQLConnection) ExecuteCommand(command string) ([]byte, error) {
	rows, err := c.DB.Query(command)
	if err != nil {
		return nil, fmt.Errorf("Error executing query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice for column values
		values := make([]interface{}, len(columns))
		// Create references for scan
		references := make([]interface{}, len(columns))
		for i := range values {
			references[i] = &values[i]
		}

		// Scan the row values into references
		if err := rows.Scan(references...); err != nil {
			return nil, err
		}

		// Build a map of column -> value for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		results = append(results, rowMap)
	}

	// Convert results to JSON
	data, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *PostgreSQLConnection) Type() string {
	return "database"
}

func (c *PostgreSQLConnection) SubType() string {
	return "postgresql"
}

func (c *PostgreSQLConnection) CommandType() string {
	return "psql"
}
