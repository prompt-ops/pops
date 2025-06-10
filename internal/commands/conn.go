package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prompt-ops/pops/internal/ui"
	"github.com/prompt-ops/pops/pkg/config"
)

// NewConnCmd creates the conn command with subcommands
func NewConnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conn",
		Short: "Manage connections to databases, clusters, and cloud resources",
		Long: `Create, list, open, and delete connections to various infrastructure resources.
Supported connection types: database, kubernetes, cloud`,
	}

	cmd.AddCommand(
		NewConnCreateCmd(),
		NewConnListCmd(),
		NewConnOpenCmd(),
		NewConnDeleteCmd(),
	)

	return cmd
}

// NewConnCreateCmd creates the conn create command
func NewConnCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new connection",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Connection creation not yet implemented.")
			fmt.Println("Use the configuration file at ~/.pops/connections.json")
		},
	}
}

// NewConnListCmd creates the conn list command  
func NewConnListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all connections",
		Run: func(cmd *cobra.Command, args []string) {
			// Simple command-line version for now
			listConnections()
		},
	}
}

// NewConnOpenCmd creates the conn open command
func NewConnOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [connection-name]",
		Short: "Open a connection shell",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			
			if len(args) == 0 {
				// List connections and let user choose
				listConnectionsForSelection()
			} else {
				// Open specific connection
				openConnection(args[0], dryRun)
			}
		},
	}
	
	cmd.Flags().Bool("dry-run", false, "Test connection and shell setup without opening interactive shell")
	return cmd
}

// NewConnDeleteCmd creates the conn delete command
func NewConnDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [connection-name]",
		Short: "Delete a connection",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Connection deletion not yet implemented for: %s\n", args[0])
		},
	}
}

// Implementation functions

func listConnections() {
	connections, err := config.GetAllConnections()
	if err != nil {
		fmt.Printf("Error getting connections: %v\n", err)
		return
	}

	if len(connections) == 0 {
		fmt.Println("No connections found.")
		fmt.Println("Create connections in ~/.pops/connections.json")
		return
	}

	fmt.Printf("Found %d connection(s):\n\n", len(connections))
	for i, conn := range connections {
		fmt.Printf("%d. %s (%s/%s)\n", 
			i+1, 
			conn.Name, 
			conn.Type.GetMainType(), 
			conn.Type.GetSubtype())
	}
}

func listConnectionsForSelection() {
	connections, err := config.GetAllConnections()
	if err != nil {
		fmt.Printf("Error getting connections: %v\n", err)
		return
	}

	if len(connections) == 0 {
		fmt.Println("No connections found.")
		fmt.Println("Create connections in ~/.pops/connections.json")
		return
	}

	fmt.Println("Available connections:")
	for i, conn := range connections {
		fmt.Printf("%d. %s (%s/%s)\n", 
			i+1, 
			conn.Name, 
			conn.Type.GetMainType(), 
			conn.Type.GetSubtype())
	}
	fmt.Println("\nUse: pops conn open <connection-name>")
}

func openConnection(connectionName string, dryRun bool) {
	conn, err := config.GetConnectionByName(connectionName)
	if err != nil {
		fmt.Printf("Error: Connection '%s' not found\n", connectionName)
		fmt.Println("Available connections:")
		listConnections()
		return
	}

	fmt.Printf("Opening connection: %s (%s/%s)\n", 
		conn.Name, 
		conn.Type.GetMainType(), 
		conn.Type.GetSubtype())
	
	// Convert old connection to domain connection
	domainConn := ui.ConvertToDomainConnection(conn)
	
	// Create shell service
	shellService := ui.CreateShellService(domainConn)
	
	if dryRun {
		fmt.Println("✅ Dry run mode - connection and shell service created successfully")
		fmt.Println("Connection details:")
		fmt.Printf("  Name: %s\n", conn.Name)
		fmt.Printf("  Type: %s\n", conn.Type.GetMainType())
		fmt.Printf("  Subtype: %s\n", conn.Type.GetSubtype())
		fmt.Printf("  Domain Type: %s\n", domainConn.Type)
		fmt.Printf("  Status: %s\n", domainConn.Status)
		
		// Test shell service
		fmt.Println("\n🧪 Testing shell service...")
		display, err := shellService.ShowContext(context.Background(), connectionName)
		if err != nil {
			fmt.Printf("❌ Context fetch failed: %v\n", err)
		} else {
			fmt.Printf("✅ Context loaded: %s\n", display.Title)
			fmt.Printf("   Items: %d\n", len(display.Items))
		}
		return
	}
	
	// Check if we can open a TTY (interactive terminal)
	if !isTerminalAvailable() {
		fmt.Println("Interactive shell requires a terminal.")
		fmt.Println("Use --dry-run flag to test without opening interactive shell.")
		fmt.Println("Connection details:")
		fmt.Printf("  Name: %s\n", conn.Name)
		fmt.Printf("  Type: %s\n", conn.Type.GetMainType())
		fmt.Printf("  Subtype: %s\n", conn.Type.GetSubtype())
		return
	}

	// Create and run the interactive shell
	shell := ui.NewSimpleShell(connectionName, shellService)
	program := tea.NewProgram(shell, tea.WithAltScreen())
	
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running shell: %v\n", err)
	}
}

func isTerminalAvailable() bool {
	// Check if we have a TTY available
	if _, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		return false
	}
	return true
}