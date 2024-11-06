package connection

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
	"github.com/olekukonko/tablewriter"
	"github.com/peterh/liner"
)

type RDBMSConnection struct {
	ConnectionString string
	DB               *sql.DB
}

func NewRDBMSConnection(connectionString string) (*RDBMSConnection, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Error pinging the database: %v", err)
	}

	return &RDBMSConnection{
		ConnectionString: connectionString,
		DB:               db,
	}, nil
}

func handleRDBMSConnection(name string) {
	reader := bufio.NewReader(os.Stdin)
	color.Cyan("Enter the RDBMS connection string: ")
	connectionString, _ := reader.ReadString('\n')
	connectionString = strings.TrimSpace(connectionString)

	// Save the connection details
	conn := Connection{
		Type: "rdbms",
		Name: name,
	}
	if err := SaveConnection(conn); err != nil {
		color.Red("Error saving connection: %v", err)
		return
	}

	color.Green("Creating RDBMS connection '%s' with connection string '%s'", name, connectionString)

	rdbmsConn, err := NewRDBMSConnection(connectionString)
	if err != nil {
		color.Red("Error creating RDBMS connection: %v", err)
		return
	}

	tables, err := rdbmsConn.GetTables()
	if err != nil {
		color.Red("Error getting tables: %v", err)
		return
	}

	if len(tables) > 0 {
		color.Green("Tables in the database:")
		for _, table := range tables {
			color.Yellow("- " + table)
		}
	} else {
		color.Yellow("No tables found in the database.")
	}

	color.Cyan("Starting **pops** interactive shell. Type your SQL query, or type 'exit' to quit.")

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	historyFile := filepath.Join(os.TempDir(), ".pops_history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		prompt := fmt.Sprintf("[%s] > ", name)
		input, err := line.Prompt(prompt)
		if err == liner.ErrPromptAborted {
			color.Cyan("Exiting PromptOps shell.")
			break
		} else if err != nil {
			color.Red("Error reading line: %s", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		line.AppendHistory(input)

		if input == "exit" {
			color.Cyan("Exiting PromptOps shell.")
			break
		}

		// Generate the SQL query using OpenAI
		sqlQuery, err := getCommand(input, RDBMSQuery)
		if err != nil {
			color.Red("Error generating SQL query: %s", err)
			continue
		}

		color.Red("Query: %s", sqlQuery)

		// Execute the SQL query
		result, err := rdbmsConn.ExecuteQuery(sqlQuery)
		if err != nil {
			color.Red("Error executing query: %s", err)
			continue
		}

		color.Green("Query result:")
		fmt.Println(result)
	}

	if f, err := os.Create(historyFile); err != nil {
		color.Red("Error writing history file: %s", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

func (r *RDBMSConnection) GetTables() ([]string, error) {
	rows, err := r.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return nil, fmt.Errorf("Error querying tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("Error scanning table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

func (r *RDBMSConnection) ExecuteQuery(query string) (string, error) {
	rows, err := r.DB.Query(query)
	if err != nil {
		return "", fmt.Errorf("Error executing query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("Error getting columns: %v", err)
	}

	// Initialize table writer with buffer
	var tableOutput strings.Builder
	table := tablewriter.NewWriter(&tableOutput)
	table.SetHeader(columns)

	// Prepare a slice for each column
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate over rows
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("Error scanning row: %v", err)
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					row[i] = string(v)
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		table.Append(row)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("Row iteration error: %v", err)
	}

	// Render the table to the buffer
	table.Render()

	return tableOutput.String(), nil
}
