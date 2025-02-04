package conn

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
	"github.com/prompt-ops/pops/pkg/ai"
)

var (
	// TODO: Rename?

	// PostgreSQLDatabaseConnection
	PostgreSQLDatabaseConnection = AvailableDatabaseConnectionType{
		Subtype: "PostgreSQL",
		Driver:  "postgres",
	}

	// MySQLDatabaseConnection
	MySQLDatabaseConnection = AvailableDatabaseConnectionType{
		Subtype: "MySQL",
		Driver:  "mysql",
	}

	// MongoDBDatabaseConnection
	MongoDBDatabaseConnection = AvailableDatabaseConnectionType{
		Subtype: "MongoDB",
		Driver:  "mongodb",
	}

	// AvailableDatabaseConnectionTypes is a list of available database connections.
	AvailableDatabaseConnectionTypes = []AvailableDatabaseConnectionType{
		PostgreSQLDatabaseConnection,
		MySQLDatabaseConnection,
		MongoDBDatabaseConnection,
	}
)

// AvailableDatabaseConnection is a helper struct to UI to list available database connection types.
// Subtype will be shown in the UI.
// Driver will be saved in the connection details.
type AvailableDatabaseConnectionType struct {
	Subtype string
	Driver  string
}

type DatabaseConnectionType struct {
	// MainType of the connection type.
	// Example: "database".
	MainType string `json:"mainType"`

	// Subtype is the subtype of the database connection type.
	// Example: "PostgreSQL", "MySQL", "MongoDB".
	Subtype string `json:"subtype"`
}

func (d DatabaseConnectionType) GetMainType() string {
	return "Database"
}

func (d DatabaseConnectionType) GetSubtype() string {
	return d.Subtype
}

type DatabaseConnectionDetails struct {
	// ConnectionString is the connection string for the database.
	ConnectionString string `json:"connectionString"`

	// Driver is the driver name for the database.
	// Example: "postgres", "mysql", "mongodb".
	Driver string `json:"driver"`
}

func (d DatabaseConnectionDetails) GetDriver() string {
	return d.Driver
}

func (d DatabaseConnectionDetails) GetConnectionString() string {
	return d.ConnectionString
}

// NewDatabaseConnection creates a new database connection.
func NewDatabaseConnection(name string, availableDatabaseConnectionType AvailableDatabaseConnectionType, connectionString string) Connection {
	return Connection{
		Name: name,
		Type: DatabaseConnectionType{
			MainType: "Database",
			Subtype:  availableDatabaseConnectionType.Subtype,
		},
		Details: DatabaseConnectionDetails{
			ConnectionString: connectionString,
			Driver:           availableDatabaseConnectionType.Driver,
		},
	}
}

// GetDatabaseConnectionDetails retrieves the DatabaseConnectionDetails from a Connection.
func GetDatabaseConnectionDetails(conn Connection) (DatabaseConnectionDetails, error) {
	if conn.Type.GetMainType() != ConnectionTypeDatabase {
		return DatabaseConnectionDetails{}, fmt.Errorf("connection is not of type 'database'")
	}
	details, ok := conn.Details.(DatabaseConnectionDetails)
	if !ok {
		return DatabaseConnectionDetails{}, fmt.Errorf("invalid connection details for 'database'")
	}
	return details, nil
}

// BaseDatabaseConnection is a partial implementation of the ConnectionInterface for databases.
type BaseDatabaseConnection struct {
	Connection Connection
}

func (d *BaseDatabaseConnection) GetConnection() Connection {
	return d.Connection
}

type BaseRDBMSConnection struct {
	BaseDatabaseConnection

	// TablesAndColumns is a map of tables and their columns.
	// This will be set via SetContext.
	TablesAndColumns map[string][]ColumnDetail

	// DB is the database connection.
	DB *sql.DB
}

func (b *BaseRDBMSConnection) CheckAuthentication() error {
	connectionDetails, err := GetDatabaseConnectionDetails(b.Connection)
	if err != nil {
		return err
	}

	db, err := sql.Open(connectionDetails.Driver, connectionDetails.ConnectionString)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging the database: %v", err)
	}

	return nil
}

// SetContext sets the context for the RDBMS connection.
// It gets the tables and their columns.
func (b *BaseRDBMSConnection) SetContext() error {
	connectionDetails, err := GetDatabaseConnectionDetails(b.Connection)
	if err != nil {
		return err
	}

	db, err := sql.Open(connectionDetails.Driver, connectionDetails.ConnectionString)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	defer db.Close()

	query, ok := TablesAndColumnsQueryMap[connectionDetails.Driver]
	if !ok {
		return fmt.Errorf("unsupported driver: %s", connectionDetails.Driver)
	}

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying database schema: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, table, column, dataType string
		if err := rows.Scan(&schema, &table, &column, &dataType); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		schema = AddQuotesIfNeeded(schema)
		table = AddQuotesIfNeeded(table)
		column = AddQuotesIfNeeded(column)
		dataType = AddQuotesIfNeeded(dataType)

		fullTableName := fmt.Sprintf(`%s.%s`, schema, table)
		b.TablesAndColumns[fullTableName] = append(b.TablesAndColumns[fullTableName], ColumnDetail{
			Name:     column,
			DataType: dataType,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %v", err)
	}

	return nil
}

// AddQuotesIfNeeded adds quotes around the name if it contains capital letters.
func AddQuotesIfNeeded(name string) string {
	// if regexp.MustCompile(`[A-Z]`).MatchString(name) {
	// 	return fmt.Sprintf(`"%s"`, name)
	// }
	// return name
	return fmt.Sprintf(`"%s"`, name)
}

// GetContext returns the tables and columns set by SetContext.
func (b *BaseRDBMSConnection) GetContext() string {
	if b.TablesAndColumns == nil {
		// Call SetContext to populate the tables and columns.
		// This is a fallback in case SetContext is not called.
		if err := b.SetContext(); err != nil {
			return fmt.Sprintf("Error getting context: %v", err)
		}
	}

	context := fmt.Sprintf("%s Connection Details:\n", b.Connection.Type.GetSubtype())
	context += "Note to the AI: Please use all columns and table with double quotes as defined below.\n"
	context += "Note to the AI: And please always use tables with aliases where possible.\n"
	context += "Database Schema:\n"

	// If still no tables found, return an error message.
	if len(b.TablesAndColumns) == 0 {
		context += "No tables found or SetContext() not called.\n"
		return context
	}

	// Iterate over each table and its columns
	for table, columns := range b.TablesAndColumns {
		context += fmt.Sprintf("- **%s**:\n", table)
		for _, column := range columns {
			context += fmt.Sprintf("  - `%s` (%s)\n", column.Name, column.DataType)
		}
	}

	return context
}

// GetFormattedContext generates a pretty-printed string of the tables and columns.
func (b *BaseRDBMSConnection) GetFormattedContext() (string, error) {
	if b.TablesAndColumns == nil {
		// Call SetContext to populate the tables and columns.
		if err := b.SetContext(); err != nil {
			return "", fmt.Errorf("error getting context: %v", err)
		}
	}

	if len(b.TablesAndColumns) == 0 {
		return "No tables found or SetContext() not called.", nil
	}

	var buffer bytes.Buffer
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Margin(1, 0)

	columnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212"))

	for tableName, columns := range b.TablesAndColumns {
		var tableBuffer bytes.Buffer
		tableBuffer.WriteString(fmt.Sprintf("Table: %s\n", tableName))
		tableBuffer.WriteString("Columns:\n")
		for _, column := range columns {
			columnContent := fmt.Sprintf("%s (%s)\n", column.Name, column.DataType)
			tableBuffer.WriteString(columnStyle.Render(columnContent))
		}
		tableContent := tableBuffer.String()
		buffer.WriteString(tableStyle.Render(tableContent))
		buffer.WriteString("\n")
	}

	return buffer.String(), nil
}

func (b *BaseRDBMSConnection) ExecuteCommand(command string) ([]byte, error) {
	connectionDetails, err := GetDatabaseConnectionDetails(b.Connection)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(connectionDetails.Driver, connectionDetails.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(command)
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

	data, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *BaseRDBMSConnection) FormatResultAsTable(result []byte) (string, error) {
	var rows []map[string]interface{}
	if err := json.Unmarshal(result, &rows); err != nil {
		return "", fmt.Errorf("failed to parse JSON result: %v", err)
	}

	if len(rows) == 0 {
		return "No data available", nil
	}
	var header []string
	for col := range rows[0] {
		header = append(header, col)
	}

	var tableRows [][]string
	for _, row := range rows {
		var tableRow []string
		for _, col := range header {
			val := row[col]
			tableRow = append(tableRow, fmt.Sprintf("%v", val))
		}
		tableRows = append(tableRows, tableRow)
	}

	var buffer bytes.Buffer
	table := tablewriter.NewWriter(&buffer)
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)
	table.SetRowLine(true)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(header)

	for _, row := range tableRows {
		table.Append(row)
	}
	table.Render()

	return buffer.String(), nil
}

type PostgreSQLConnection struct {
	BaseRDBMSConnection
}

var _ ConnectionInterface = &PostgreSQLConnection{}

func NewPostgreSQLConnection(connnection *Connection) *PostgreSQLConnection {
	if connnection.Type.GetSubtype() != "PostgreSQL" {
		panic("Connection type is not PostgreSQL")
	}

	return &PostgreSQLConnection{
		BaseRDBMSConnection{
			BaseDatabaseConnection{
				Connection: *connnection,
			},
			map[string][]ColumnDetail{},
			nil,
		},
	}
}

func (p *PostgreSQLConnection) GetCommand(prompt string) (string, error) {
	if p.TablesAndColumns == nil {
		// Call SetContext to populate the tables and columns.
		// This is a fallback in case SetContext is not called.
		if err := p.SetContext(); err != nil {
			return "", fmt.Errorf("Error getting command: %v", err)
		}
	}

	aiModel, err := ai.NewOpenAIModel(p.CommandType(), p.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	cmd, err := aiModel.GetCommand(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get command from AI: %v", err)
	}

	return cmd.Command, nil
}

func (p *PostgreSQLConnection) GetAnswer(prompt string) (string, error) {
	if p.TablesAndColumns == nil {
		// Call SetContext to populate the tables and columns.
		// This is a fallback in case SetContext is not called.
		if err := p.SetContext(); err != nil {
			return "", fmt.Errorf("Error getting answer: %v", err)
		}
	}

	aiModel, err := ai.NewOpenAIModel(p.CommandType(), p.GetContext())
	if err != nil {
		return "", fmt.Errorf("failed to create AI model: %v", err)
	}

	answer, err := aiModel.GetAnswer(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get answer from AI: %v", err)
	}

	return answer.Answer, nil
}

func (p *PostgreSQLConnection) CommandType() string {
	return "psql"
}

// ColumnDetail is a helper struct to store the column details.
// For now, it only stores the name and the data type.
type ColumnDetail struct {
	Name     string
	DataType string
}
