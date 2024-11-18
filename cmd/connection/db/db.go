package db

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	_ "github.com/lib/pq"              // Import the PostgreSQL driver
	"github.com/peterh/liner"

	config "github.com/prompt-ops/cli/cmd/config"
)

func HandleDatabaseConnection(name string) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt the user to select a database driver
	color.Cyan("Select a database driver:")
	color.Cyan("1. PostgreSQL")
	color.Cyan("2. MySQL")
	color.Cyan("3. MongoDB")
	color.Cyan("Enter the number of the driver you want to use: ")
	driverSelection, _ := reader.ReadString('\n')
	driverSelection = strings.TrimSpace(driverSelection)

	var conn DatabaseConnection
	switch driverSelection {
	case "1":
		conn = NewPostgresConnection("")
	case "2":
		conn = NewMySQLConnection("")
	case "3":
		conn = NewMongoDBConnection("")
	default:
		color.Red("Invalid selection")
		return
	}

	color.Cyan("Enter the database connection string: ")
	connectionString, _ := reader.ReadString('\n')
	connectionString = strings.TrimSpace(connectionString)

	// Set the connection string
	switch driverSelection {
	case "1":
		conn = NewPostgresConnection(connectionString)
	case "2":
		conn = NewMySQLConnection(connectionString)
	case "3":
		conn = NewMongoDBConnection(connectionString)
	}

	// Connect to the database
	if err := conn.Connect(); err != nil {
		color.Red("Error connecting to the database: %v", err)
		return
	}
	defer conn.Disconnect()

	// Save the connection details
	connection := config.Connection{
		Type: "db",
		Name: name,
	}
	if err := config.SaveConnection(connection); err != nil {
		color.Red("Error saving connection: %v", err)
		return
	}

	color.Green("Creating a database connection '%s' with connection string '%s'", name, connectionString)

	tables, err := conn.GetTables()
	if err != nil {
		color.Red("Error getting tables: %v", err)
		return
	}

	if len(tables) > 0 {
		color.Green("Tables/Collections in the database:")
		for _, table := range tables {
			color.Yellow("- " + table)
		}
	} else {
		color.Yellow("No tables found in the database.")
	}

	// Fetch additional database information
	tableColumns := make(map[string]map[string]string)
	for _, table := range tables {
		columns, err := conn.GetTableColumns(table)
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
		parsedResponse, err := getCommand(input, conn.GetType(), dbContext)
		if err != nil {
			color.Red("Error generating SQL query: %s", err)
			continue
		}

		color.Red("Query: %s", parsedResponse.Command)

		// Execute the SQL query
		result, err := conn.ExecuteQuery(parsedResponse.Command)
		if err != nil {
			color.Red("Error executing query: %s", err)
			continue
		}

		color.Green("Query result:")
		fmt.Println(result)

		// Display suggested next steps
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

		// 		result, err = conn.ExecuteQuery(parsedResponse.Command)
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
