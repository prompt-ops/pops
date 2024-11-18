package db

type DatabaseType struct {
	Type    string
	Command string
}

type DatabaseConnection interface {
	Connect() error
	Disconnect() error
	GetTables() ([]string, error)
	GetTableColumns(tableName string) (map[string]string, error)
	ExecuteQuery(query string) (string, error)
	GetType() DatabaseType
}
