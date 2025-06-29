package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/c-bata/go-prompt"
)

type (
	ShellContext struct {
		Client        driver.Client
		DB            driver.Database
		CurrentDB     string
		Context       context.Context
		Config        *ShellConfig
		ConnectionURL string
		ConfigManager *ConfigManager
		CurrentConfig string
	}
	ShellConfig struct {
		Host     string
		Port     int
		Username string
		UseSSL   bool
		Password string
		DBName   string
	}
)

func NewShellContext(config *ShellConfig) (*ShellContext, error) {
	protocol := "http"
	if config.UseSSL {
		protocol = "https"
	}

	connectionURL := fmt.Sprintf("%s://%s:%d", protocol, config.Host, config.Port)
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{connectionURL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.Username, config.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	ctx := context.Background()
	dbName := config.DBName
	if dbName == "" {
		dbName = "_system"
	}

	db, err := client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %v", dbName, err)
	}

	return &ShellContext{
		Client:        client,
		DB:            db,
		CurrentDB:     dbName,
		Context:       ctx,
		Config:        config,
		ConnectionURL: connectionURL,
	}, nil
}

func NewShellContextWithConfig(config *ShellConfig, configManager *ConfigManager, configName string) (*ShellContext, error) {
	shellCtx, err := NewShellContext(config)
	if err != nil {
		return nil, err
	}

	shellCtx.ConfigManager = configManager
	shellCtx.CurrentConfig = configName
	return shellCtx, nil
}

func (s *ShellContext) showCurrentConnection() {
	fmt.Printf("Current connection:\n")
	fmt.Printf("  Config: %s\n", s.CurrentConfig)
	fmt.Printf("  Host: %s\n", s.Config.Host)
	fmt.Printf("  Port: %d\n", s.Config.Port)
	fmt.Printf("  Database: %s\n", s.CurrentDB)
	fmt.Printf("  Username: %s\n", s.Config.Username)
	fmt.Printf("  SSL: %t\n", s.Config.UseSSL)
	fmt.Printf("  URL: %s\n", s.ConnectionURL)
}

func startShell(s *ShellContext) {
	promptPrefix := func() string {
		if s.CurrentConfig != "manual" {
			return fmt.Sprintf("arango[%s:%s]> ", s.CurrentConfig, s.CurrentDB)
		}
		return fmt.Sprintf("arango[%s]> ", s.CurrentDB)
	}
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

			if strings.HasPrefix(strings.TrimSpace(input), "/") {
				s.handleSpecialCommands(input)
				return
			}

			s.executor(input)
		},
		completer,
		prompt.OptionPrefix(promptPrefix()),
		prompt.OptionTitle("ArangoDB Shell"),
	)
	p.Run()
}
