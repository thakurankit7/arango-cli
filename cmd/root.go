package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "arango-cli",
	Short: "A CLI tool for ArangoDB",
	Long:  `A command line interface to interact with ArangoDB databases and collections.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be defined here
}
