package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var (
	host     string
	port     int
	username string
	password string
	dbName   string
	useSSL   bool
)

var shellCmd = &cobra.Command{
	Use:   "arango",
	Short: "Start an interactive ArangoDB shell",
	Long:  `Connect to ArangoDB and start an interactive shell similar to the mysql client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		protocol := "http"
		if useSSL {
			protocol = "https"
		}

		connectionURL := fmt.Sprintf("%s://%s:%d", protocol, host, port)

		conn, err := http.NewConnection(http.ConnectionConfig{
			Endpoints: []string{connectionURL},
		})
		if err != nil {
			return fmt.Errorf("failed to create connection: %v", err)
		}

		client, err := driver.NewClient(driver.ClientConfig{
			Connection:     conn,
			Authentication: driver.BasicAuthentication(username, password),
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}

		ctx := context.Background()
		var db driver.Database

		if dbName == "" {
			dbName = "_system"
		}

		db, err = client.Database(ctx, dbName)
		if err != nil {
			return fmt.Errorf("failed to connect to database '%s': %v", dbName, err)
		}

		fmt.Printf("Connected to ArangoDB at %s, database: %s\n", connectionURL, dbName)
		fmt.Println("Type 'help' for help, 'exit' to quit")

		// Start interactive shell
		startShell(ctx, db, client)
		return nil
	},
}

func completer(d prompt.Document) []prompt.Suggest {
	// AQL keyword suggestions here
	s := []prompt.Suggest{
		{Text: "show collections", Description: "List collections"},
		{Text: "col", Description: "List collections (shorthand)"},
		{Text: "show databases", Description: "List databases"},
		{Text: "db", Description: "List databases (shorthand)"},
		{Text: "use", Description: "Switch database"},
		{Text: "FOR", Description: "AQL FOR loop"},
		{Text: "RETURN", Description: "AQL RETURN statement"},
		{Text: "FILTER", Description: "AQL FILTER statement"},
		{Text: "SORT", Description: "AQL SORT statement"},
		{Text: "LIMIT", Description: "AQL LIMIT statement"},
		{Text: "LET", Description: "AQL variable assignment"},
		{Text: "COLLECT", Description: "AQL COLLECT statement"},
		{Text: "INSERT", Description: "AQL INSERT statement"},
		{Text: "UPDATE", Description: "AQL UPDATE statement"},
		{Text: "REPLACE", Description: "AQL REPLACE statement"},
		{Text: "REMOVE", Description: "AQL REMOVE statement"},
		{Text: "exit", Description: "Exit the shell"},
		{Text: "quit", Description: "Exit the shell"},
		{Text: "help", Description: "Show help"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func startShell(ctx context.Context, db driver.Database, client driver.Client) {
	p := prompt.New(
		func(input string) {
			input = strings.TrimSpace(input)
			if input == "" {
				return
			}

			switch strings.ToLower(input) {
			case "exit", "quit":
				fmt.Println("Goodbye!")
				os.Exit(0)
			case "help":
				printHelp()
				return
			}

			if handleSpecialCommands(ctx, db, client, input) {
				return
			}

			executeQuery(ctx, db, input)
		},
		completer,
		prompt.OptionPrefix("arango> "),
		prompt.OptionTitle("ArangoDB Shell"),
	)
	p.Run()
}

func handleSpecialCommands(ctx context.Context, db driver.Database, client driver.Client, input string) bool {
	lowerInput := strings.ToLower(input)

	switch true {
	case lowerInput == "show databases" || lowerInput == "db":
		showDatabases(ctx, client)
		return true
	case lowerInput == "show collections" || lowerInput == "col":
		showCollections(ctx, db)
		return true
	case strings.HasPrefix(lowerInput, "use "):
		useDatabase(ctx, db, client, strings.TrimPrefix(input, "use "))
		return true
	default:
		fmt.Printf("Unknown command: %s\n", input)
	}
	return false
}

func showDatabases(ctx context.Context, client driver.Client) {
	dbs, err := client.Databases(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Databases:")
	for i, d := range dbs {
		name := d.Name()
		fmt.Printf("%d: %s\n", i+1, name)
	}
}

func showCollections(ctx context.Context, db driver.Database) {
	cols, err := db.Collections(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Collections:")
	for i, col := range cols {
		fmt.Printf("%d: %s\n", i+1, col.Name())
	}
}

func useDatabase(ctx context.Context, db driver.Database, client driver.Client, dbName string) {
	if len(dbName) < 2 {
		fmt.Println("Usage: :use <database>")
		return
	}

	newDb, err := client.Database(ctx, dbName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	db = newDb
	fmt.Printf("Using database '%s'\n", dbName)
}

func executeQuery(ctx context.Context, db driver.Database, query string) {
	if !strings.Contains(strings.ToUpper(query), "RETURN") &&
		!strings.Contains(strings.ToUpper(query), "INSERT") &&
		!strings.Contains(strings.ToUpper(query), "UPDATE") &&
		!strings.Contains(strings.ToUpper(query), "REMOVE") &&
		!strings.Contains(strings.ToUpper(query), "REPLACE") {
		query = "RETURN " + query
	}

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cursor.Close()

	fmt.Println("Results:")
	var result interface{}
	count := 0
	for {
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			fmt.Printf("Error reading result: %v\n", err)
			return
		}
		fmt.Printf("%d: %v\n", count, result)
		count++
	}
	fmt.Printf("%d record(s) returned\n", count)
}

func printHelp() {
	help := `
	ArangoDB Shell Commands:
	:collections, :col          List collections in current database
	:databases, :db             List available databases
	:use <database>             Switch to a different database
	exit, quit                  Exit the shell
	help                        Display this help message

	Any other input will be executed as an AQL query.
	Example queries:
	RETURN DOCUMENT("users/123")
	FOR doc IN users RETURN doc
	`
	fmt.Println(help)
}

func init() {
	rootCmd.AddCommand(shellCmd)

	// Reuse the same flags from the connect command
	shellCmd.Flags().StringVarP(&host, "host", "H", "localhost", "ArangoDB host")
	shellCmd.Flags().IntVarP(&port, "port", "p", 8529, "ArangoDB port")
	shellCmd.Flags().StringVarP(&username, "username", "u", "root", "ArangoDB username")
	shellCmd.Flags().StringVarP(&password, "password", "P", "", "ArangoDB password")
	shellCmd.Flags().StringVarP(&dbName, "database", "d", "_system", "Database name to connect to")
	shellCmd.Flags().BoolVarP(&useSSL, "ssl", "s", false, "Use SSL for connection")

	shellCmd.MarkFlagRequired("password")
}
