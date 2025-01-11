package connection

var TablesAndColumnsQueryMap = map[string]string{
	"postgres": `
        SELECT table_schema, table_name, column_name, data_type
        FROM information_schema.columns
        WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
        ORDER BY table_schema, table_name, ordinal_position;
    `,
	"mysql": `
        SELECT table_schema, table_name, column_name, data_type
        FROM information_schema.columns
        WHERE table_schema NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
        ORDER BY table_schema, table_name, ordinal_position;
    `,
}
