package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSL      bool   `yaml:"ssl"`
}

type Config struct {
	Databases map[string]DatabaseConfig `yaml:"databases"`
	Default   string                    `yaml:"default"`
}

type ConfigManager struct {
	config     *Config
	configPath string
}

func NewConfigManager() (*ConfigManager, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(currentDir, "config", ".env.yaml")

	cm := &ConfigManager{
		configPath: configPath,
	}

	if err := cm.loadConfig(); err != nil {
		return nil, err
	}

	return cm, nil
}

func (cm *ConfigManager) loadConfig() error {
	data, err := ioutil.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cm.createDefaultConfig()
		}
		return err
	}

	cm.config = &Config{}
	if err := yaml.Unmarshal(data, cm.config); err != nil {
		return err
	}

	return nil
}

func (cm *ConfigManager) createDefaultConfig() error {
	defaultConfig := &Config{
		Databases: map[string]DatabaseConfig{
			"local": {
				Host:     "localhost",
				Port:     8529,
				Username: "root",
				Password: "",
				Database: "_system",
				SSL:      false,
			},
		},
		Default: "local",
	}

	cm.config = defaultConfig
	return cm.saveConfig()
}

func (cm *ConfigManager) saveConfig() error {
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return err
	}

	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(cm.configPath, data, 0644)
}

func (s *ShellContext) listConfigs() {
	if s.ConfigManager == nil {
		fmt.Println("Configuration manager not available")
		return
	}

	configs := s.ConfigManager.ListDatabases()
	defaultConfig := s.ConfigManager.GetDefaultDatabase()

	fmt.Println("Available database configurations:")
	for _, name := range configs {
		marker := "  "
		if name == defaultConfig {
			marker = "* "
		}
		if name == s.CurrentConfig {
			marker = "→ "
		}

		config, _ := s.ConfigManager.GetDatabaseConfig(name)
		fmt.Printf("%s%s: %s:%d/%s\n", marker, name, config.Host, config.Port, config.Database)
	}

	if len(configs) == 0 {
		fmt.Println("  No database configurations found.")
		fmt.Println("  Create ~/.arango-cli/env.yaml to add configurations.")
	}
}

func (s *ShellContext) switchToConfig(configName string) {
	if s.ConfigManager == nil {
		fmt.Println("Configuration manager not available")
		return
	}

	dbConfig, err := s.ConfigManager.GetDatabaseConfig(configName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Switching to configuration '%s' (%s:%d)...\n", configName, dbConfig.Host, dbConfig.Port)

	newShellConfig := &ShellConfig{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		Username: dbConfig.Username,
		Password: dbConfig.Password,
		UseSSL:   dbConfig.SSL,
		DBName:   dbConfig.Database,
	}

	newShellCtx, err := NewShellContext(newShellConfig)
	if err != nil {
		fmt.Printf("Failed to connect to '%s': %v\n", configName, err)
		return
	}

	s.Client = newShellCtx.Client
	s.DB = newShellCtx.DB
	s.CurrentDB = newShellCtx.CurrentDB
	s.Config = newShellCtx.Config
	s.ConnectionURL = newShellCtx.ConnectionURL
	s.CurrentConfig = configName

	fmt.Printf("Successfully switched to '%s' (database: %s)\n", configName, s.CurrentDB)
}

func (cm *ConfigManager) GetDatabaseConfig(name string) (DatabaseConfig, error) {
	if config, exists := cm.config.Databases[name]; exists {
		return config, nil
	}
	return DatabaseConfig{}, fmt.Errorf("database config '%s' not found", name)
}

func (cm *ConfigManager) ListDatabases() []string {
	var names []string
	for name := range cm.config.Databases {
		names = append(names, name)
	}
	return names
}

func (cm *ConfigManager) GetDefaultDatabase() string {
	return cm.config.Default
}
