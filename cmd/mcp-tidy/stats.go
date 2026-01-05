package main

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/nnnkkk7/mcp-tidy/config"
	"github.com/nnnkkk7/mcp-tidy/transcript"
	"github.com/nnnkkk7/mcp-tidy/types"
	"github.com/nnnkkk7/mcp-tidy/ui"
	"github.com/spf13/cobra"
)

var (
	statsPeriod string
	statsJSON   bool
	statsSort   string
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show MCP server usage statistics",
	Long: `Display usage statistics for MCP servers based on Claude Code transcript logs.

Shows call counts, last used time, and a visual usage bar for each server.
Servers that haven't been used in the specified period are marked as unused.`,
	RunE: runStats,
}

func init() {
	statsCmd.Flags().StringVar(&statsPeriod, "period", "30d", "Time period (7d, 30d, 90d, all)")
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "Output in JSON format")
	statsCmd.Flags().StringVar(&statsSort, "sort", "calls", "Sort order (calls, name, last-used)")
}

func runStats(_ *cobra.Command, _ []string) error {
	transcriptPath := transcript.DefaultTranscriptPath()
	configPath := config.DefaultConfigPath()
	period := types.ParsePeriod(statsPeriod)

	// Get usage stats from transcript logs
	stats, err := transcript.GetStats(transcriptPath, period)
	if err != nil {
		return err
	}

	// Load configured servers
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Merge: add configured servers that have no stats (0 calls)
	stats = mergeConfiguredServers(stats, cfg.Servers())

	// Sort stats
	sortStats(stats, statsSort)

	if statsJSON {
		return outputStatsJSON(stats)
	}

	ui.RenderStatsTable(os.Stdout, stats, period.Duration())
	return nil
}

// mergeConfiguredServers adds configured servers that don't appear in stats
func mergeConfiguredServers(stats []types.ServerStats, servers []types.MCPServer) []types.ServerStats {
	// Create a map of existing stats by name
	statsMap := make(map[string]bool)
	for _, s := range stats {
		statsMap[s.Name] = true
	}

	// Add configured servers that don't have stats
	for i := range servers {
		if !statsMap[servers[i].Name] {
			stats = append(stats, types.ServerStats{
				Name:  servers[i].Name,
				Calls: 0,
				// LastUsed is zero value (never used)
			})
			statsMap[servers[i].Name] = true // prevent duplicates
		}
	}

	return stats
}

func sortStats(stats []types.ServerStats, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Name < stats[j].Name
		})
	case "last-used":
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].LastUsed.After(stats[j].LastUsed)
		})
	default: // "calls"
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Calls > stats[j].Calls
		})
	}
}

type statsOutput struct {
	Servers    []serverStatsOutput `json:"servers"`
	TotalCalls int                 `json:"totalCalls"`
	Period     string              `json:"period"`
}

type serverStatsOutput struct {
	Name     string `json:"name"`
	Calls    int    `json:"calls"`
	LastUsed string `json:"lastUsed"`
	Unused   bool   `json:"unused"`
}

func outputStatsJSON(stats []types.ServerStats) error {
	output := statsOutput{
		Period:  statsPeriod,
		Servers: make([]serverStatsOutput, len(stats)),
	}

	period := types.ParsePeriod(statsPeriod)
	for i, s := range stats {
		output.TotalCalls += s.Calls
		lastUsed := "never"
		if !s.LastUsed.IsZero() {
			lastUsed = s.LastUsed.Format("2006-01-02T15:04:05Z07:00")
		}
		output.Servers[i] = serverStatsOutput{
			Name:     s.Name,
			Calls:    s.Calls,
			LastUsed: lastUsed,
			Unused:   s.IsUnused(period.Duration()),
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
