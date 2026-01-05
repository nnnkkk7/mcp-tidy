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
	removeUnused  bool
	removeDryRun  bool
	removeForce   bool
	removePeriod  string
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

func runRemove(cmd *cobra.Command, args []string) error {
	configPath := config.DefaultConfigPath()

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	servers := cfg.Servers()
	if len(servers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	// Get stats for all servers
	transcriptPath := transcript.DefaultTranscriptPath()
	period := types.ParsePeriod(removePeriod)
	allStats, err := transcript.GetStats(transcriptPath, period)
	if err != nil {
		// Stats might fail if no logs exist, continue anyway
		allStats = nil
	}

	// Convert stats to map
	statsMap := make(map[string]types.ServerStats)
	for _, s := range allStats {
		statsMap[s.Name] = s
	}

	// Filter servers if --unused
	displayServers := servers
	if removeUnused {
		var unused []types.MCPServer
		for _, s := range servers {
			stat, ok := statsMap[s.Name]
			if !ok || stat.IsUnused(period.Duration()) {
				unused = append(unused, s)
			}
		}
		displayServers = unused

		if len(displayServers) == 0 {
			fmt.Println("No unused servers found.")
			return nil
		}
	}

	// Let user select servers to remove
	selectedIdx := ui.SelectServersPrompt(displayServers, statsMap)
	if len(selectedIdx) == 0 {
		fmt.Println("No servers selected.")
		return nil
	}

	// Get selected servers
	var toRemove []types.MCPServer
	for _, idx := range selectedIdx {
		if idx >= 0 && idx < len(displayServers) {
			toRemove = append(toRemove, displayServers[idx])
		}
	}

	// Dry run mode
	if removeDryRun {
		ui.RenderDryRunSummary(os.Stdout, toRemove)
		return nil
	}

	// Confirm removal
	if !removeForce {
		prompt := fmt.Sprintf("Remove %d server(s)?", len(toRemove))
		if !ui.ConfirmPrompt(prompt, false) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove servers
	if err := config.RemoveServers(configPath, toRemove); err != nil {
		return err
	}

	ui.RenderRemovalSummary(os.Stdout, toRemove)
	return nil
}
