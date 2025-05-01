package cmd

import "fmt"

func printHelp() {
	help := `
	ArangoDB Shell Commands:
	/collections, /col          List collections in current database
	/databases, /db             List available databases
	/use <database>             Switch to a different database
	exit, quit                  Exit the shell
	help                        Display this help message

	Any other input will be executed as an AQL query.
	Example queries:
	RETURN DOCUMENT("users/123")
	FOR doc IN users RETURN doc
	`
	fmt.Println(help)
}
