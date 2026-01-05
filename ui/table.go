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
func RenderStatsTable(w io.Writer, stats []types.ServerStats, period time.Duration) {
	if len(stats) == 0 {
		fmt.Fprintln(w, "No usage data found.")
		return
	}

	// Sort by calls (descending)
	sorted := make([]types.ServerStats, len(stats))
	copy(sorted, stats)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Calls > sorted[j].Calls
	})

	// Find max calls for bar scaling
	maxCalls := 0
	totalCalls := 0
	for _, s := range sorted {
		if s.Calls > maxCalls {
			maxCalls = s.Calls
		}
		totalCalls += s.Calls
	}

	periodDays := int(period.Hours() / 24)
	fmt.Fprintf(w, "\nMCP Server Usage Statistics (last %d days)\n", periodDays)
	fmt.Fprintln(w, strings.Repeat("─", tableWidth))
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

	fmt.Fprintf(w, "\nTotal tool calls: %d\n\n", totalCalls)
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
