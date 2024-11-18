package connection

import "fmt"

func handleDatabaseConnection(args []string) {
	fmt.Println("Handling Database connection")
	for i, arg := range args {
		fmt.Printf("arg[%d]: %s\n", i, arg)
	}
}
