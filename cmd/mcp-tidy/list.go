package main

import (
	"os"

	"github.com/nnnkkk7/mcp-tidy/config"
	"github.com/nnnkkk7/mcp-tidy/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Long: `Display all MCP servers configured in ~/.claude.json.

Shows both global servers and project-specific servers with their
scope, type, and command/URL.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	configPath := config.DefaultConfigPath()

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	servers := cfg.Servers()
	ui.RenderServerTable(os.Stdout, servers)

	return nil
}
