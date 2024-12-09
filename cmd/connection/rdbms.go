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

	config "github.com/prompt-ops/cli/cmd/config"
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
	conn := config.Connection{
		Type: "rdbms",
		Name: name,
	}
	if err := config.SaveConnection(conn); err != nil {
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

	// Fetch additional database information
	tableColumns := make(map[string]map[string]string)
	for _, table := range tables {
		columns, err := rdbmsConn.GetTableColumns(table)
		if err != nil {
			color.Red("Error getting columns for table %s: %v", table, err)
			continue
		}
		tableColumns[table] = columns
	}

	// Provide context to the AI
	dbContext := fmt.Sprintf("Tables: %v, Columns: %v", tables, tableColumns)

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
		parsedResponse, err := getCommand(input, RDBMSQuery, dbContext)
		if err != nil {
			color.Red("Error generating SQL query: %s", err)
			continue
		}

		color.Red("Query: %s", parsedResponse.Command)

		// Execute the SQL query
		result, err := rdbmsConn.ExecuteQuery(parsedResponse.Command)
		if err != nil {
			color.Red("Error executing query: %s", err)
			continue
		}

		color.Green("Query result:")
		fmt.Println(result)

		// // Display suggested next steps
		// if len(parsedResponse.SuggestedSteps) > 0 {
		// 	nextStep, err := selectNextStep(parsedResponse.SuggestedSteps)
		// 	if err != nil {
		// 		color.Red("Error: %s", err)
		// 		continue
		// 	}

		// 	if nextStep != "" {
		// 		color.Green("\nExecuting selected step: %s", nextStep)
		// 		parsedResponse, err = getCommand(nextStep, RDBMSQuery, dbContext)
		// 		if err != nil {
		// 			color.Red("Error processing selected step: %s", err)
		// 			continue
		// 		}

		// 		result, err = rdbmsConn.ExecuteQuery(parsedResponse.Command)
		// 		if err != nil {
		// 			color.Red("Error executing query: %s", err)
		// 			continue
		// 		}

		// 		color.Green("Query result:")
		// 		fmt.Println(result)
		// 	} else {
		// 		color.Yellow("Skipping suggested steps.")
		// 	}
		// } else {
		// 	color.Yellow("No suggested next steps available.")
		// }
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

func (r *RDBMSConnection) GetTableColumns(tableName string) (map[string]string, error) {
	query := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", tableName)
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Error querying columns for table %s: %v", tableName, err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return nil, fmt.Errorf("Error scanning column for table %s: %v", tableName, err)
		}
		columns[columnName] = dataType
	}

	return columns, nil
}
