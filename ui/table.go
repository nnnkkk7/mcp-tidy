// Package ui provides terminal UI components for mcp-tidy.
package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nnnkkk7/mcp-tidy/types"
)

const (
	barWidth     = 16
	barFilled    = "█"
	barEmpty     = "░"
	tableWidth   = 80
	maxPathWidth = 35
)

var (
	warningColor = color.New(color.FgYellow)
	successColor = color.New(color.FgGreen)
	dimColor     = color.New(color.Faint)
)

// RenderServerTable renders a table of MCP servers.
func RenderServerTable(w io.Writer, servers []types.MCPServer) {
	if len(servers) == 0 {
		fmt.Fprintln(w, "No MCP servers configured.")
		return
	}

	// Sort servers by name
	sorted := make([]types.MCPServer, len(servers))
	copy(sorted, servers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	fmt.Fprintf(w, "\nMCP Servers (%d configured)\n", len(servers))
	fmt.Fprintln(w, strings.Repeat("─", tableWidth))
	fmt.Fprintf(w, "  %-14s %-36s %s\n", "NAME", "SCOPE", "COMMAND")

	for i := range sorted {
		scope := sorted[i].ScopeString()
		if len(scope) > maxPathWidth {
			scope = "..." + scope[len(scope)-maxPathWidth+3:]
		}

		command := sorted[i].CommandString()
		if len(command) > 40 {
			command = command[:37] + "..."
		}

		fmt.Fprintf(w, "  %-14s %-36s %s\n", sorted[i].Name, scope, command)
	}
	fmt.Fprintln(w)
}

// RenderStatsTable renders a table of server usage statistics.
// If servers is provided, stats are grouped by scope (global/project).
func RenderStatsTable(w io.Writer, stats []types.ServerStats, period time.Duration, servers ...[]types.MCPServer) {
	if len(stats) == 0 {
		fmt.Fprintln(w, "No usage data found.")
		return
	}

	// Build stats map for quick lookup
	statsMap := make(map[string]types.ServerStats)
	for _, s := range stats {
		statsMap[s.Name] = s
	}

	// Find max calls for bar scaling
	maxCalls := 0
	totalCalls := 0
	for _, s := range stats {
		if s.Calls > maxCalls {
			maxCalls = s.Calls
		}
		totalCalls += s.Calls
	}

	periodDays := int(period.Hours() / 24)
	fmt.Fprintf(w, "\nMCP Server Usage Statistics (last %d days)\n", periodDays)
	fmt.Fprintln(w, strings.Repeat("─", tableWidth))

	// If servers provided, render grouped by scope
	if len(servers) > 0 && len(servers[0]) > 0 {
		renderGroupedStats(w, servers[0], statsMap, maxCalls, period)
	} else {
		// Fallback to simple list (backwards compatibility)
		renderSimpleStats(w, stats, maxCalls, period)
	}

	fmt.Fprintf(w, "\nTotal tool calls: %d\n\n", totalCalls)
}

// renderGroupedStats renders stats grouped by scope (global/project).
func renderGroupedStats(w io.Writer, servers []types.MCPServer, statsMap map[string]types.ServerStats, maxCalls int, period time.Duration) {
	// Separate servers by scope
	var globalServers []types.MCPServer
	projectGroups := make(map[string][]types.MCPServer)

	for i := range servers {
		if servers[i].Scope == types.ScopeGlobal {
			globalServers = append(globalServers, servers[i])
		} else {
			projectGroups[servers[i].ProjectPath] = append(projectGroups[servers[i].ProjectPath], servers[i])
		}
	}

	// Render global servers
	if len(globalServers) > 0 {
		fmt.Fprintf(w, "\n%s\n", dimColor.Sprint("── Global ──"))
		fmt.Fprintf(w, "  %-14s %6s   %-14s %s\n", "NAME", "CALLS", "LAST USED", "USAGE")
		renderServerStatsRows(w, globalServers, statsMap, maxCalls, period)
	}

	// Render project servers grouped by project
	if len(projectGroups) > 0 {
		// Get sorted project paths
		var projectPaths []string
		for path := range projectGroups {
			projectPaths = append(projectPaths, path)
		}
		sort.Strings(projectPaths)

		for _, projectPath := range projectPaths {
			projectServers := projectGroups[projectPath]
			displayPath := projectPath
			if len(displayPath) > 50 {
				displayPath = "..." + displayPath[len(displayPath)-47:]
			}
			fmt.Fprintf(w, "\n%s\n", dimColor.Sprintf("── %s ──", displayPath))
			fmt.Fprintf(w, "  %-14s %6s   %-14s %s\n", "NAME", "CALLS", "LAST USED", "USAGE")
			renderServerStatsRows(w, projectServers, statsMap, maxCalls, period)
		}
	}
}

// renderServerStatsRows renders stats rows for a list of servers.
func renderServerStatsRows(w io.Writer, servers []types.MCPServer, statsMap map[string]types.ServerStats, maxCalls int, period time.Duration) {
	// Sort by calls (descending)
	sorted := make([]types.MCPServer, len(servers))
	copy(sorted, servers)
	sort.Slice(sorted, func(i, j int) bool {
		si := statsMap[sorted[i].Name]
		sj := statsMap[sorted[j].Name]
		return si.Calls > sj.Calls
	})

	for i := range sorted {
		stat, ok := statsMap[sorted[i].Name]
		if !ok {
			stat = types.ServerStats{Name: sorted[i].Name}
		}

		bar := RenderUsageBar(stat.Calls, maxCalls, barWidth)
		lastUsed := stat.LastUsedString()

		line := fmt.Sprintf("  %-14s %6d   %-14s %s", stat.Name, stat.Calls, lastUsed, bar)

		if stat.IsUnused(period) {
			fmt.Fprintf(w, "%s  %s\n", line, warningColor.Sprint("⚠️ unused"))
		} else {
			fmt.Fprintln(w, line)
		}
	}
}

// renderSimpleStats renders stats as a simple list (backwards compatibility).
func renderSimpleStats(w io.Writer, stats []types.ServerStats, maxCalls int, period time.Duration) {
	// Sort by calls (descending)
	sorted := make([]types.ServerStats, len(stats))
	copy(sorted, stats)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Calls > sorted[j].Calls
	})

	fmt.Fprintf(w, "  %-14s %6s   %-14s %s\n", "NAME", "CALLS", "LAST USED", "USAGE")

	for _, s := range sorted {
		bar := RenderUsageBar(s.Calls, maxCalls, barWidth)
		lastUsed := s.LastUsedString()

		line := fmt.Sprintf("  %-14s %6d   %-14s %s", s.Name, s.Calls, lastUsed, bar)

		if s.IsUnused(period) {
			fmt.Fprintf(w, "%s  %s\n", line, warningColor.Sprint("⚠️ unused"))
		} else {
			fmt.Fprintln(w, line)
		}
	}
}

// RenderUsageBar renders a usage bar with filled and empty portions.
func RenderUsageBar(calls, maxCalls, width int) string {
	if maxCalls == 0 {
		return strings.Repeat(barEmpty, width)
	}

	filled := (calls * width) / maxCalls
	if filled > width {
		filled = width
	}

	filledPart := strings.Repeat(barFilled, filled)
	emptyPart := strings.Repeat(barEmpty, width-filled)

	return successColor.Sprint(filledPart) + dimColor.Sprint(emptyPart)
}

// RenderRemovalSummary renders a summary of removed servers.
func RenderRemovalSummary(w io.Writer, removed []types.MCPServer) {
	if len(removed) == 0 {
		fmt.Fprintln(w, "No servers removed.")
		return
	}

	fmt.Fprintln(w)
	for i := range removed {
		location := "~/.claude.json"
		if removed[i].Scope == types.ScopeProject {
			location = fmt.Sprintf("~/.claude.json (project: %s)", removed[i].ProjectPath)
		}
		successColor.Fprintf(w, "✓ Removed: %s (from %s)\n", removed[i].Name, location)
	}
	fmt.Fprintln(w)
}

// RenderDryRunSummary renders a dry-run summary of what would be removed.
func RenderDryRunSummary(w io.Writer, servers []types.MCPServer) {
	if len(servers) == 0 {
		fmt.Fprintln(w, "No servers would be removed.")
		return
	}

	fmt.Fprintln(w, "\n[DRY RUN] The following servers would be removed:")
	for i := range servers {
		location := "global"
		if servers[i].Scope == types.ScopeProject {
			location = servers[i].ProjectPath
		}
		fmt.Fprintf(w, "  - %s (%s)\n", servers[i].Name, location)
	}
	fmt.Fprintln(w, "\nRun without --dry-run to actually remove these servers.")
}
