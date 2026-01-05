// Package main provides the entry point for mcp-tidy CLI.
package main

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mcp-tidy",
	Short: "Manage Claude Code MCP servers",
	Long: `mcp-tidy helps you visualize MCP server usage and remove unused servers.

Like 'go mod tidy', it helps keep your MCP configuration clean and organized.`,
	Version: Version,
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(removeCmd)
}
