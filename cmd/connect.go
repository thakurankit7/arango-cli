package cmd

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var (
	host            string
	port            int
	username        string
	password        string
	dbName          string
	useSSL          bool
	buffer          strings.Builder
	isMultilineMode bool
	configName      string
)

var shellCmd = &cobra.Command{
	Use:   "arango",
	Short: "Start an interactive ArangoDB shell",
	Long:  `Connect to ArangoDB and start an interactive shell similar to the mysql client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configManager, err := NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %v", err)
		}

		var config *ShellConfig
		var currentConfigName string
		if configName != "" {
			// Connect using saved configuration
			dbConfig, err := configManager.GetDatabaseConfig(configName)
			if err != nil {
				return fmt.Errorf("failed to get database config '%s': %v", configName, err)
			}

			config = &ShellConfig{
				Host:     dbConfig.Host,
				Port:     dbConfig.Port,
				Username: dbConfig.Username,
				Password: dbConfig.Password,
				UseSSL:   dbConfig.SSL,
				DBName:   dbConfig.Database,
			}
			currentConfigName = configName
		} else {
			// Connect using command line flags
			config = &ShellConfig{
				Host:     host,
				Port:     port,
				Username: username,
				Password: password,
				UseSSL:   useSSL,
				DBName:   dbName,
			}
			currentConfigName = "manual"
		}

		shellCtx, err := NewShellContextWithConfig(config, configManager, currentConfigName)
		if err != nil {
			return fmt.Errorf("failed to initialize shell: %v", err)
		}

		fmt.Printf("Connected to ArangoDB at %s, database: %s\n", shellCtx.ConnectionURL, shellCtx.CurrentDB)
		fmt.Println("Type 'help' for help, 'exit' to quit")

		// Start interactive shell
		startShell(shellCtx)
		return nil
	},
}

func completer(d prompt.Document) []prompt.Suggest {
	// AQL keyword suggestions here
	s := []prompt.Suggest{
		{Text: "/show collections", Description: "List collections"},
		{Text: "/col", Description: "List collections (shorthand)"},
		{Text: "/show databases", Description: "List databases"},
		{Text: "/db", Description: "List databases (shorthand)"},
		{Text: "/use", Description: "Switch database"},
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

func (s *ShellContext) handleSpecialCommands(input string) bool {
	lowerInput := strings.ToLower(input)
	parts := strings.Fields(input)

	switch true {
	case lowerInput == "/show databases" || lowerInput == "/db":
		showDatabases(s.Context, s.Client)
		return true
	case lowerInput == "/show collections" || lowerInput == "/col":
		showCollections(s.Context, s.DB)
		return true
	case strings.HasPrefix(lowerInput, "/use "):
		s.useDatabase(strings.TrimPrefix(input, "/use "))
		return true
	case lowerInput == "/list configs" || lowerInput == "/configs":
		s.listConfigs()
		return true
	case strings.HasPrefix(lowerInput, "/switch "):
		if len(parts) >= 2 {
			s.switchToConfig(parts[1])
		} else {
			fmt.Println("Usage: /switch <config_name>")
		}
		return true
	case lowerInput == "/current":
		s.showCurrentConnection()
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

func (s *ShellContext) useDatabase(dbName string) {
	fmt.Println("Switching database...", dbName)
	if len(dbName) < 2 {
		fmt.Println("Usage: :use <database>")
		return
	}

	newDb, err := s.Client.Database(s.Context, dbName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	s.DB = newDb
	s.CurrentDB = dbName
	fmt.Printf("Using database '%s'\n", dbName)
}

func (s *ShellContext) executor(input string) {
	// Check if line ends with a semi colon (force execution)
	if strings.HasSuffix(strings.TrimSpace(input), ";") {
		trimmedInput := strings.TrimSpace(input)
		trimmedInput = trimmedInput[:len(trimmedInput)-1]

		if isMultilineMode {
			// Finalize the multiline query and execute
			fullQuery := buffer.String() + trimmedInput
			buffer.Reset()
			isMultilineMode = false
			s.executeQuery(fullQuery)
		} else {
			// Execute single line query (without the semi colon)
			s.executeQuery(trimmedInput)
		}
		return
	}

	// If we're already in multiline mode, add this line to the buffer
	if isMultilineMode {
		buffer.WriteString(input + "\n")
		return
	}

	// start multiline mode
	buffer.WriteString(input + "\n")
	isMultilineMode = true
}

func (s *ShellContext) executeQuery(query string) {
	if !strings.Contains(strings.ToUpper(query), "RETURN") &&
		!strings.Contains(strings.ToUpper(query), "INSERT") &&
		!strings.Contains(strings.ToUpper(query), "UPDATE") &&
		!strings.Contains(strings.ToUpper(query), "REMOVE") &&
		!strings.Contains(strings.ToUpper(query), "REPLACE") {
		query = "RETURN " + query
	}

	cursor, err := s.DB.Query(s.Context, query, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cursor.Close()

	var resultData []interface{}
	count := 0
	for {
		var doc interface{}
		_, err := cursor.ReadDocument(s.Context, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			ShowPopup(fmt.Sprintf("Error reading result: %v", err))
			return
		}
		resultData = append(resultData, doc)
		count++
	}
	stats := cursor.Statistics()

	// Format the data and show in popup
	formattedData := FormatQueryResult(resultData, stats)
	ShowPopup(formattedData)
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
