package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nnnkkk7/mcp-tidy/config"
	"github.com/nnnkkk7/mcp-tidy/transcript"
	"github.com/nnnkkk7/mcp-tidy/types"
	"github.com/nnnkkk7/mcp-tidy/ui"
)

// Integration tests for mcp-tidy commands

func TestListCommand_NoServers(t *testing.T) {
	cfg, err := config.Load("../../testdata/claude_empty.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	servers := cfg.Servers()
	var buf bytes.Buffer
	ui.RenderServerTable(&buf, servers)

	output := buf.String()
	if !strings.Contains(output, "No MCP servers configured") {
		t.Errorf("expected 'No MCP servers configured' message, got: %s", output)
	}
}

func TestListCommand_WithServers(t *testing.T) {
	cfg, err := config.Load("../../testdata/claude.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	servers := cfg.Servers()
	var buf bytes.Buffer
	ui.RenderServerTable(&buf, servers)

	output := buf.String()

	// Check expected servers are present
	expectedServers := []string{"context7", "serena"}
	for _, name := range expectedServers {
		if !strings.Contains(output, name) {
			t.Errorf("expected server %q in output, got: %s", name, output)
		}
	}

	// Check scope information is present
	if !strings.Contains(output, "global") {
		t.Errorf("expected 'global' scope in output, got: %s", output)
	}
}

func TestStatsCommand_NoLogs(t *testing.T) {
	// Create empty temp directory
	tmpDir := t.TempDir()

	stats, err := transcript.GetStats(tmpDir, types.Period30Days)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d entries", len(stats))
	}

	var buf bytes.Buffer
	ui.RenderStatsTable(&buf, stats, 30*24*time.Hour)

	output := buf.String()
	if !strings.Contains(output, "No usage data") {
		t.Errorf("expected 'No usage data' message, got: %s", output)
	}
}

func TestStatsCommand_WithStats(t *testing.T) {
	stats, err := transcript.GetStats("../../testdata/projects", types.PeriodAll)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	// testdata/projects/-Users-xxx-github-proj/session1.jsonl contains:
	// - mcp__context7__resolve-library-id (1 call)
	// - mcp__context7__query-docs (1 call)
	// - mcp__serena__find_symbol (1 call)

	if len(stats) < 2 {
		t.Fatalf("expected at least 2 server stats, got %d", len(stats))
	}

	// Check context7 stats
	var context7Stats *types.ServerStats
	var serenaStats *types.ServerStats
	for i := range stats {
		switch stats[i].Name {
		case "context7":
			context7Stats = &stats[i]
		case "serena":
			serenaStats = &stats[i]
		}
	}

	if context7Stats == nil {
		t.Error("expected context7 stats to be present")
	} else if context7Stats.Calls < 2 {
		t.Errorf("expected context7 to have at least 2 calls, got %d", context7Stats.Calls)
	}

	if serenaStats == nil {
		t.Error("expected serena stats to be present")
	} else if serenaStats.Calls < 1 {
		t.Errorf("expected serena to have at least 1 call, got %d", serenaStats.Calls)
	}
}

func TestStatsCommand_JSONOutput(t *testing.T) {
	stats, err := transcript.GetStats("../../testdata/projects", types.PeriodAll)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	// Create JSON output like the command does
	output := statsOutput{
		Period:  "all",
		Servers: make([]serverStatsOutput, len(stats)),
	}

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
			Unused:   s.IsUnused(types.PeriodAll.Duration()),
		}
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Verify JSON is valid and contains expected fields
	var decoded statsOutput
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if decoded.Period != "all" {
		t.Errorf("expected period 'all', got %q", decoded.Period)
	}
	if len(decoded.Servers) != len(stats) {
		t.Errorf("expected %d servers, got %d", len(stats), len(decoded.Servers))
	}
}

func TestStatsCommand_PeriodFilter(t *testing.T) {
	tests := []struct {
		name   string
		period types.Period
	}{
		{"7 days", types.Period7Days},
		{"30 days", types.Period30Days},
		{"90 days", types.Period90Days},
		{"all time", types.PeriodAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := transcript.GetStats("../../testdata/projects", tt.period)
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			// Stats should be filtered by period
			// All testdata entries have old timestamps, so recent periods may have fewer results
			var buf bytes.Buffer
			ui.RenderStatsTable(&buf, stats, tt.period.Duration())

			// Output should not panic and should be valid
			if buf.Len() == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestRemoveCommand_DryRun(t *testing.T) {
	// Test dry-run mode by loading config and simulating removal
	cfg, err := config.Load("../../testdata/claude.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	servers := cfg.Servers()
	if len(servers) == 0 {
		t.Fatal("expected servers in config")
	}

	// Simulate selecting first server for removal
	toRemove := []types.MCPServer{servers[0]}

	// Test dry-run rendering
	var buf bytes.Buffer
	ui.RenderDryRunSummary(&buf, toRemove)

	output := buf.String()
	if !strings.Contains(output, "DRY RUN") || !strings.Contains(output, servers[0].Name) {
		t.Errorf("dry-run output should contain 'DRY RUN' and server name, got: %s", output)
	}
}

func TestRemoveCommand_Force(t *testing.T) {
	// Create a temporary copy of the config file
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "claude.json")

	// Copy testdata config
	srcContent, err := os.ReadFile("../../testdata/claude.json")
	if err != nil {
		t.Fatalf("failed to read source config: %v", err)
	}
	if err := os.WriteFile(tmpConfig, srcContent, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Load config
	cfg, err := config.Load(tmpConfig)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	initialCount := len(cfg.Servers())
	if initialCount == 0 {
		t.Fatal("expected servers in config")
	}

	// Remove first server
	toRemove := []types.MCPServer{cfg.Servers()[0]}

	// Actually remove (force mode - no confirmation needed)
	if err := config.RemoveServers(tmpConfig, toRemove); err != nil {
		t.Fatalf("failed to remove servers: %v", err)
	}

	// Verify removal
	cfg2, err := config.Load(tmpConfig)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if len(cfg2.Servers()) != initialCount-1 {
		t.Errorf("expected %d servers after removal, got %d", initialCount-1, len(cfg2.Servers()))
	}
}

func TestRemoveCommand_Unused(t *testing.T) {
	cfg, err := config.Load("../../testdata/claude.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	servers := cfg.Servers()
	if len(servers) == 0 {
		t.Fatal("expected servers in config")
	}

	// Get stats for filtering
	stats, _ := transcript.GetStats("../../testdata/projects", types.Period30Days)
	statsMap := make(map[string]types.ServerStats)
	for _, s := range stats {
		statsMap[s.Name] = s
	}

	// Filter unused servers (servers with no stats or unused according to period)
	period := types.Period30Days
	var unused []types.MCPServer
	for _, server := range servers {
		stat, ok := statsMap[server.Name]
		if !ok || stat.IsUnused(period.Duration()) {
			unused = append(unused, server)
		}
	}

	// Verify filtering logic works
	// Note: testdata logs have old timestamps, so servers should appear as unused
	t.Logf("Total servers: %d, Unused: %d", len(servers), len(unused))
}

func TestSortStats(t *testing.T) {
	stats := []types.ServerStats{
		{Name: "zebra", Calls: 10, LastUsed: time.Now().Add(-1 * time.Hour)},
		{Name: "alpha", Calls: 100, LastUsed: time.Now().Add(-24 * time.Hour)},
		{Name: "beta", Calls: 50, LastUsed: time.Now()},
	}

	tests := []struct {
		name      string
		sortBy    string
		wantFirst string
	}{
		{"sort by calls (default)", "calls", "alpha"},
		{"sort by name", "name", "alpha"},
		{"sort by last-used", "last-used", "beta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid mutating original
			statsCopy := make([]types.ServerStats, len(stats))
			copy(statsCopy, stats)

			sortStats(statsCopy, tt.sortBy)

			if statsCopy[0].Name != tt.wantFirst {
				t.Errorf("sortStats(%q) first element = %q, want %q",
					tt.sortBy, statsCopy[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestStatsOutput_JSON(t *testing.T) {
	now := time.Now()
	stats := []types.ServerStats{
		{Name: "context7", Calls: 100, LastUsed: now},
		{Name: "serena", Calls: 50, LastUsed: now.Add(-24 * time.Hour)},
	}

	output := statsOutput{
		Period:  "30d",
		Servers: make([]serverStatsOutput, len(stats)),
	}

	for i, s := range stats {
		output.TotalCalls += s.Calls
		lastUsed := s.LastUsed.Format("2006-01-02T15:04:05Z07:00")
		output.Servers[i] = serverStatsOutput{
			Name:     s.Name,
			Calls:    s.Calls,
			LastUsed: lastUsed,
			Unused:   s.IsUnused(30 * 24 * time.Hour),
		}
	}

	wantTotalCalls := 150
	if diff := cmp.Diff(wantTotalCalls, output.TotalCalls); diff != "" {
		t.Errorf("TotalCalls mismatch (-want +got):\n%s", diff)
	}

	if output.Period != "30d" {
		t.Errorf("expected Period '30d', got %q", output.Period)
	}

	if len(output.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(output.Servers))
	}
}

func TestMergeConfiguredServers(t *testing.T) {
	tests := []struct {
		name        string
		stats       []types.ServerStats
		servers     []types.MCPServer
		wantCount   int
		wantServers []string
	}{
		{
			name: "adds configured server with no stats",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 10},
			},
			servers: []types.MCPServer{
				{Name: "context7"},
				{Name: "serena"},
				{Name: "puppeteer"},
			},
			wantCount:   3,
			wantServers: []string{"context7", "serena", "puppeteer"},
		},
		{
			name:  "all configured servers have no stats",
			stats: []types.ServerStats{},
			servers: []types.MCPServer{
				{Name: "new-server"},
				{Name: "another-server"},
			},
			wantCount:   2,
			wantServers: []string{"new-server", "another-server"},
		},
		{
			name: "no new servers to add",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 10},
				{Name: "serena", Calls: 5},
			},
			servers: []types.MCPServer{
				{Name: "context7"},
				{Name: "serena"},
			},
			wantCount:   2,
			wantServers: []string{"context7", "serena"},
		},
		{
			name: "duplicate server names in config are handled",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 10},
			},
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "context7", Scope: types.ScopeProject}, // same name, different scope
				{Name: "serena"},
			},
			wantCount:   2, // context7 appears once, serena once
			wantServers: []string{"context7", "serena"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeConfiguredServers(tt.stats, tt.servers)

			if len(result) != tt.wantCount {
				t.Errorf("mergeConfiguredServers() returned %d servers, want %d", len(result), tt.wantCount)
			}

			// Check all expected servers are present
			serverNames := make(map[string]bool)
			for _, s := range result {
				serverNames[s.Name] = true
			}

			for _, name := range tt.wantServers {
				if !serverNames[name] {
					t.Errorf("expected server %q not found in result", name)
				}
			}
		})
	}
}

func TestMergeConfiguredServers_ZeroCalls(t *testing.T) {
	// Verify that newly added servers have 0 calls
	stats := []types.ServerStats{
		{Name: "context7", Calls: 100},
	}
	servers := []types.MCPServer{
		{Name: "context7"},
		{Name: "unused-server"},
	}

	result := mergeConfiguredServers(stats, servers)

	for _, s := range result {
		if s.Name == "unused-server" {
			if s.Calls != 0 {
				t.Errorf("unused-server should have 0 calls, got %d", s.Calls)
			}
			if !s.LastUsed.IsZero() {
				t.Errorf("unused-server should have zero LastUsed, got %v", s.LastUsed)
			}
			return
		}
	}
	t.Error("unused-server not found in result")
}

func TestFilterServersForRemoval(t *testing.T) {
	now := time.Now()
	period := types.Period30Days

	tests := []struct {
		name         string
		servers      []types.MCPServer
		statsMap     map[string]types.ServerStats
		removeUnused bool
		wantCount    int
		wantNames    []string
	}{
		{
			name: "returns all servers when removeUnused is false",
			servers: []types.MCPServer{
				{Name: "used", Scope: types.ScopeGlobal},
				{Name: "unused", Scope: types.ScopeGlobal},
			},
			statsMap: map[string]types.ServerStats{
				"used":   {Name: "used", Calls: 100, LastUsed: now},
				"unused": {Name: "unused", Calls: 0},
			},
			removeUnused: false,
			wantCount:    2,
			wantNames:    []string{"used", "unused"},
		},
		{
			name: "filters only unused servers when removeUnused is true",
			servers: []types.MCPServer{
				{Name: "used", Scope: types.ScopeGlobal},
				{Name: "unused-no-stats", Scope: types.ScopeGlobal},
				{Name: "unused-zero-calls", Scope: types.ScopeGlobal},
			},
			statsMap: map[string]types.ServerStats{
				"used":              {Name: "used", Calls: 100, LastUsed: now},
				"unused-zero-calls": {Name: "unused-zero-calls", Calls: 0, LastUsed: time.Time{}},
			},
			removeUnused: true,
			wantCount:    2,
			wantNames:    []string{"unused-no-stats", "unused-zero-calls"},
		},
		{
			name: "filters old usage as unused",
			servers: []types.MCPServer{
				{Name: "recent", Scope: types.ScopeGlobal},
				{Name: "old", Scope: types.ScopeGlobal},
			},
			statsMap: map[string]types.ServerStats{
				"recent": {Name: "recent", Calls: 100, LastUsed: now.Add(-1 * 24 * time.Hour)}, // 1 day ago
				"old":    {Name: "old", Calls: 50, LastUsed: now.Add(-60 * 24 * time.Hour)},    // 60 days ago
			},
			removeUnused: true,
			wantCount:    1,
			wantNames:    []string{"old"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flag
			removeUnused = tt.removeUnused

			result := filterServersForRemoval(tt.servers, tt.statsMap, period)

			// For removeUnused=false, it returns input as-is
			if !tt.removeUnused {
				if len(result) != len(tt.servers) {
					t.Errorf("expected %d servers, got %d", len(tt.servers), len(result))
				}
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("expected %d servers, got %d", tt.wantCount, len(result))
			}

			// Check expected names
			resultNames := make(map[string]bool)
			for _, s := range result {
				resultNames[s.Name] = true
			}

			for _, name := range tt.wantNames {
				if !resultNames[name] {
					t.Errorf("expected server %q in result", name)
				}
			}

			// Reset flag
			removeUnused = false
		})
	}
}

func TestStatsCommand_GroupedOutput(t *testing.T) {
	// Test that stats command with servers produces grouped output
	cfg, err := config.Load("../../testdata/claude.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	servers := cfg.Servers()
	stats, err := transcript.GetStats("../../testdata/projects", types.PeriodAll)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	// Merge stats with servers
	stats = mergeConfiguredServers(stats, servers)

	var buf bytes.Buffer
	ui.RenderStatsTable(&buf, stats, types.PeriodAll.Duration(), servers)
	output := buf.String()

	// Should have grouped output if there are both global and project servers
	hasGlobal := false
	for _, s := range servers {
		if s.Scope == types.ScopeGlobal {
			hasGlobal = true
			break
		}
	}

	if hasGlobal && !strings.Contains(output, "Global") {
		t.Error("expected 'Global' section in grouped output")
	}

	// Verify all server stats are shown
	for _, s := range stats {
		if !strings.Contains(output, s.Name) {
			t.Errorf("expected server %q in output", s.Name)
		}
	}
}

func TestFilterServersForRemoval_NoUnusedServers(t *testing.T) {
	now := time.Now()
	period := types.Period30Days

	servers := []types.MCPServer{
		{Name: "active1", Scope: types.ScopeGlobal},
		{Name: "active2", Scope: types.ScopeProject, ProjectPath: "/project"},
	}

	statsMap := map[string]types.ServerStats{
		"active1": {Name: "active1", Calls: 100, LastUsed: now},
		"active2": {Name: "active2", Calls: 50, LastUsed: now},
	}

	removeUnused = true
	defer func() { removeUnused = false }()

	result := filterServersForRemoval(servers, statsMap, period)

	// Should return nil when no unused servers
	if result != nil {
		t.Errorf("expected nil when no unused servers, got %d servers", len(result))
	}
}

func TestSelectServersToRemove(t *testing.T) {
	servers := []types.MCPServer{
		{Name: "server1", Scope: types.ScopeGlobal},
		{Name: "server2", Scope: types.ScopeProject, ProjectPath: "/project"},
		{Name: "server3", Scope: types.ScopeGlobal},
	}

	statsMap := map[string]types.ServerStats{
		"server1": {Name: "server1", Calls: 100},
		"server2": {Name: "server2", Calls: 50},
		"server3": {Name: "server3", Calls: 0},
	}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantNames []string
	}{
		{
			name:      "select single server",
			input:     "1\n",
			wantCount: 1,
			wantNames: []string{"server1"},
		},
		{
			name:      "select multiple servers",
			input:     "1 3\n",
			wantCount: 2,
			wantNames: []string{"server1", "server3"},
		},
		{
			name:      "select all servers",
			input:     "all\n",
			wantCount: 3,
			wantNames: []string{"server1", "server2", "server3"},
		},
		{
			name:      "empty input returns nil",
			input:     "\n",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test selectServersToRemove directly since it uses stdin
			// But we can test the UI function it wraps
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}

			selectedIdx := ui.SelectServersPromptWithReader(r, w, servers, statsMap)

			if len(selectedIdx) != tt.wantCount {
				t.Errorf("expected %d selected, got %d", tt.wantCount, len(selectedIdx))
				return
			}

			if tt.wantCount == 0 {
				return
			}

			// Build result from indices
			var result []types.MCPServer
			for _, idx := range selectedIdx {
				if idx >= 0 && idx < len(servers) {
					result = append(result, servers[idx])
				}
			}

			resultNames := make(map[string]bool)
			for _, s := range result {
				resultNames[s.Name] = true
			}

			for _, name := range tt.wantNames {
				if !resultNames[name] {
					t.Errorf("expected server %q in result", name)
				}
			}
		})
	}
}
