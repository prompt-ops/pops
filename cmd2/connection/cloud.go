package connection

import "fmt"

func handleCloudConnection(args []string) {
	fmt.Println("Handling Cloud connection")
	for i, arg := range args {
		fmt.Printf("arg[%d]: %s\n", i, arg)
	}
}
