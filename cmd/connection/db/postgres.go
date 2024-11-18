package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
)

type PostgresConnection struct {
	ConnectionString string
	DB               *sql.DB
}

func NewPostgresConnection(cs string) *PostgresConnection {
	return &PostgresConnection{ConnectionString: cs}
}

func (pc *PostgresConnection) Connect() error {
	db, err := sql.Open("postgres", pc.ConnectionString)
	if err != nil {
		return fmt.Errorf("Error connecting to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Error pinging the database: %v", err)
	}

	pc.DB = db
	return nil
}

func (pc *PostgresConnection) Disconnect() error {
	if pc.DB != nil {
		return pc.DB.Close()
	}
	return nil
}

func (pc *PostgresConnection) GetTables() ([]string, error) {
	rows, err := pc.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
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

func (pc *PostgresConnection) GetTableColumns(tableName string) (map[string]string, error) {
	query := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", tableName)
	rows, err := pc.DB.Query(query)
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

// ExecuteQuery executes the given query and returns the output in a tabular format.
// TODO: We can think about if we want to return the result not in a tabular format but just as a string.
// We can have a separate formatter for the tabular format.
func (pc *PostgresConnection) ExecuteQuery(query string) (string, error) {
	rows, err := pc.DB.Query(query)
	if err != nil {
		return "", fmt.Errorf("Error executing query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("Error getting columns: %v", err)
	}

	var tableOutput strings.Builder
	table := tablewriter.NewWriter(&tableOutput)
	table.SetHeader(columns)

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

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

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("Row iteration error: %v", err)
	}

	table.Render()
	return tableOutput.String(), nil
}

func (pc *PostgresConnection) GetType() DatabaseType {
	return DatabaseType{
		Type:    "Postgres",
		Command: "psql query",
	}
}
