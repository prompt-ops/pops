package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
)

type MySQLConnection struct {
	ConnectionString string
	DB               *sql.DB
}

func NewMySQLConnection(cs string) *MySQLConnection {
	return &MySQLConnection{ConnectionString: cs}
}

func (mc *MySQLConnection) Connect() error {
	db, err := sql.Open("mysql", mc.ConnectionString)
	if err != nil {
		return fmt.Errorf("Error connecting to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Error pinging the database: %v", err)
	}

	mc.DB = db
	return nil
}

func (mc *MySQLConnection) Disconnect() error {
	if mc.DB != nil {
		return mc.DB.Close()
	}
	return nil
}

func (mc *MySQLConnection) GetTables() ([]string, error) {
	rows, err := mc.DB.Query("SHOW TABLES")
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

func (mc *MySQLConnection) GetTableColumns(tableName string) (map[string]string, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
	rows, err := mc.DB.Query(query)
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

func (mc *MySQLConnection) ExecuteQuery(query string) (string, error) {
	rows, err := mc.DB.Query(query)
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

func (mc *MySQLConnection) GetType() DatabaseType {
	return DatabaseType{
		Type:    "MySQL",
		Command: "mysql query",
	}
}
