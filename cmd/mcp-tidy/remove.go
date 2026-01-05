package main

import (
	"fmt"
	"os"

	"github.com/nnnkkk7/mcp-tidy/config"
	"github.com/nnnkkk7/mcp-tidy/transcript"
	"github.com/nnnkkk7/mcp-tidy/types"
	"github.com/nnnkkk7/mcp-tidy/ui"
	"github.com/spf13/cobra"
)

var (
	removeUnused bool
	removeDryRun bool
	removeForce  bool
	removePeriod string
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MCP servers",
	Long: `Interactively select and remove MCP servers from ~/.claude.json.

Creates a backup before making any changes. Use --dry-run to preview
changes without actually removing servers.`,
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVar(&removeUnused, "unused", false, "Only show unused servers")
	removeCmd.Flags().BoolVar(&removeDryRun, "dry-run", false, "Preview changes without removing")
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Remove without confirmation")
	removeCmd.Flags().StringVar(&removePeriod, "period", "30d", "Period for determining 'unused' (7d, 30d, 90d)")
}

func runRemove(_ *cobra.Command, _ []string) error {
	configPath := config.DefaultConfigPath()

	// Load config and stats
	servers, statsMap, period, err := loadServersWithStats(configPath)
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	// Filter servers if --unused
	displayServers := filterServersForRemoval(servers, statsMap, period)
	if displayServers == nil {
		return nil
	}

	// Let user select and get servers to remove
	toRemove := selectServersToRemove(displayServers, statsMap)
	if len(toRemove) == 0 {
		fmt.Println("No servers selected.")
		return nil
	}

	// Execute removal
	return executeRemoval(configPath, toRemove)
}

func loadServersWithStats(configPath string) ([]types.MCPServer, map[string]types.ServerStats, types.Period, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, 0, err
	}

	servers := cfg.Servers()
	period := types.ParsePeriod(removePeriod)

	// Get stats for all servers
	transcriptPath := transcript.DefaultTranscriptPath()
	allStats, err := transcript.GetStats(transcriptPath, period)
	if err != nil {
		allStats = nil
	}

	// Convert stats to map
	statsMap := make(map[string]types.ServerStats)
	for _, s := range allStats {
		statsMap[s.Name] = s
	}

	return servers, statsMap, period, nil
}

func filterServersForRemoval(servers []types.MCPServer, statsMap map[string]types.ServerStats, period types.Period) []types.MCPServer {
	if !removeUnused {
		return servers
	}

	var unused []types.MCPServer
	for i := range servers {
		stat, ok := statsMap[servers[i].Name]
		if !ok || stat.IsUnused(period.Duration()) {
			unused = append(unused, servers[i])
		}
	}

	if len(unused) == 0 {
		fmt.Println("No unused servers found.")
		return nil
	}
	return unused
}

func selectServersToRemove(displayServers []types.MCPServer, statsMap map[string]types.ServerStats) []types.MCPServer {
	selectedIdx := ui.SelectServersPrompt(displayServers, statsMap)
	if len(selectedIdx) == 0 {
		return nil
	}

	var toRemove []types.MCPServer
	for _, idx := range selectedIdx {
		if idx >= 0 && idx < len(displayServers) {
			toRemove = append(toRemove, displayServers[idx])
		}
	}
	return toRemove
}

func executeRemoval(configPath string, toRemove []types.MCPServer) error {
	if removeDryRun {
		ui.RenderDryRunSummary(os.Stdout, toRemove)
		return nil
	}

	if !removeForce {
		prompt := fmt.Sprintf("Remove %d server(s)?", len(toRemove))
		if !ui.ConfirmPrompt(prompt, false) {
			fmt.Println("Canceled.")
			return nil
		}
	}

	if err := config.RemoveServers(configPath, toRemove); err != nil {
		return err
	}

	ui.RenderRemovalSummary(os.Stdout, toRemove)
	return nil
}
