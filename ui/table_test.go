package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nnnkkk7/mcp-tidy/types"
)

func TestRenderServerTable(t *testing.T) {
	tests := []struct {
		name    string
		servers []types.MCPServer
		want    []string // substrings that should be present
	}{
		{
			name:    "empty servers",
			servers: nil,
			want:    []string{"No MCP servers configured"},
		},
		{
			name: "single global server",
			servers: []types.MCPServer{
				{
					Name:  "context7",
					Type:  types.ServerTypeHTTP,
					URL:   "https://mcp.context7.com/mcp",
					Scope: types.ScopeGlobal,
				},
			},
			want: []string{"context7", "global", "[http]"},
		},
		{
			name: "multiple servers",
			servers: []types.MCPServer{
				{
					Name:  "context7",
					Type:  types.ServerTypeHTTP,
					URL:   "https://mcp.context7.com/mcp",
					Scope: types.ScopeGlobal,
				},
				{
					Name:        "serena",
					Type:        types.ServerTypeStdio,
					Command:     "uvx",
					Args:        []string{"--from", "git+https://github.com/oraios/serena", "serena"},
					Scope:       types.ScopeProject,
					ProjectPath: "/Users/xxx/github/my-project",
				},
			},
			want: []string{"context7", "serena", "global", "/Users/xxx/github/my-project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderServerTable(&buf, tt.servers)
			output := buf.String()

			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("RenderServerTable() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestRenderStatsTable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		stats         []types.ServerStats
		servers       []types.MCPServer
		want          []string       // substrings that should be present
		notWant       []string       // substrings that should NOT be present
		minCount      map[string]int // minimum occurrence count for specific strings
		expectedOrder []string       // strings that should appear in this order
	}{
		// Simple mode tests (backwards compatible, no servers)
		{
			name:  "empty stats",
			stats: nil,
			want:  []string{"No usage data"},
		},
		{
			name: "single server stats",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 142, LastUsed: now.Add(-2 * time.Hour)},
			},
			want: []string{"context7", "142", "hours ago"},
		},
		{
			name: "unused server",
			stats: []types.ServerStats{
				{Name: "puppeteer", Calls: 0, LastUsed: time.Time{}},
			},
			want: []string{"puppeteer", "0", "never", "unused"},
		},
		{
			name: "backwards compatible without servers",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 100, LastUsed: now},
				{Name: "serena", Calls: 50, LastUsed: now},
			},
			servers: nil,
			want:    []string{"context7", "serena"},
			notWant: []string{"── Global ──"},
		},
		{
			name: "falls back to simple mode with empty servers",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 100, LastUsed: now},
			},
			servers: []types.MCPServer{},
			want:    []string{"context7"},
			notWant: []string{"── Global ──"},
		},
		// Grouped mode tests (with servers)
		{
			name: "groups global and project servers",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 100, LastUsed: now},
				{Name: "serena", Calls: 50, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path/to/project"},
			},
			want: []string{"Global", "/path/to/project", "context7", "serena", "100", "50"},
		},
		{
			name: "same server name in multiple scopes",
			stats: []types.ServerStats{
				{Name: "context7", Calls: 150, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "context7", Scope: types.ScopeProject, ProjectPath: "/project-1"},
				{Name: "context7", Scope: types.ScopeProject, ProjectPath: "/project-2"},
			},
			want:     []string{"Global", "/project-1", "/project-2"},
			minCount: map[string]int{"context7": 3},
		},
		{
			name: "truncates long project path",
			stats: []types.ServerStats{
				{Name: "serena", Calls: 50, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/Users/username/very/long/path/to/a/deeply/nested/project/directory"},
			},
			want: []string{"..."},
		},
		{
			name: "shows unused warning in grouped mode",
			stats: []types.ServerStats{
				{Name: "unused-server", Calls: 0, LastUsed: time.Time{}},
			},
			servers: []types.MCPServer{
				{Name: "unused-server", Scope: types.ScopeGlobal},
			},
			want: []string{"unused", "Global"},
		},
		{
			name: "shows server without stats",
			stats: []types.ServerStats{
				{Name: "existing", Calls: 100, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "existing", Scope: types.ScopeGlobal},
				{Name: "no-stats", Scope: types.ScopeGlobal},
			},
			want: []string{"existing", "no-stats"},
		},
		{
			name: "shows correct total calls",
			stats: []types.ServerStats{
				{Name: "server1", Calls: 100, LastUsed: now},
				{Name: "server2", Calls: 50, LastUsed: now},
				{Name: "server3", Calls: 25, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "server1", Scope: types.ScopeGlobal},
				{Name: "server2", Scope: types.ScopeProject, ProjectPath: "/project"},
				{Name: "server3", Scope: types.ScopeProject, ProjectPath: "/project"},
			},
			want: []string{"175"},
		},
		// Order tests
		{
			name: "global before projects",
			stats: []types.ServerStats{
				{Name: "global-server", Calls: 100, LastUsed: now},
				{Name: "project-a-server", Calls: 50, LastUsed: now},
				{Name: "project-b-server", Calls: 25, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "global-server", Scope: types.ScopeGlobal},
				{Name: "project-a-server", Scope: types.ScopeProject, ProjectPath: "/project-a"},
				{Name: "project-b-server", Scope: types.ScopeProject, ProjectPath: "/project-b"},
			},
			expectedOrder: []string{"Global", "/project-a", "/project-b"},
		},
		{
			name: "servers sorted by calls within group",
			stats: []types.ServerStats{
				{Name: "low-usage", Calls: 10, LastUsed: now},
				{Name: "high-usage", Calls: 100, LastUsed: now},
				{Name: "medium-usage", Calls: 50, LastUsed: now},
			},
			servers: []types.MCPServer{
				{Name: "low-usage", Scope: types.ScopeGlobal},
				{Name: "high-usage", Scope: types.ScopeGlobal},
				{Name: "medium-usage", Scope: types.ScopeGlobal},
			},
			expectedOrder: []string{"high-usage", "medium-usage", "low-usage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if tt.servers != nil {
				RenderStatsTable(&buf, tt.stats, 30*24*time.Hour, tt.servers)
			} else {
				RenderStatsTable(&buf, tt.stats, 30*24*time.Hour)
			}
			output := buf.String()

			// Check expected substrings
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}

			// Check unwanted substrings
			for _, notWant := range tt.notWant {
				if strings.Contains(output, notWant) {
					t.Errorf("output should not contain %q\nGot:\n%s", notWant, output)
				}
			}

			// Check minimum occurrence counts
			for str, minCount := range tt.minCount {
				count := strings.Count(output, str)
				if count < minCount {
					t.Errorf("expected %q at least %d times, got %d\nGot:\n%s", str, minCount, count, output)
				}
			}

			// Verify order if specified
			if len(tt.expectedOrder) > 0 {
				lastIdx := -1
				for _, str := range tt.expectedOrder {
					idx := strings.Index(output, str)
					if idx == -1 {
						t.Errorf("%q not found in output\nGot:\n%s", str, output)
						continue
					}
					if idx < lastIdx {
						t.Errorf("%q should appear after previous item\nGot:\n%s", str, output)
					}
					lastIdx = idx
				}
			}
		})
	}
}

func TestRenderUsageBar(t *testing.T) {
	tests := []struct {
		name     string
		calls    int
		maxCalls int
		width    int
		wantLen  int
	}{
		{
			name:     "full bar",
			calls:    100,
			maxCalls: 100,
			width:    16,
			wantLen:  16,
		},
		{
			name:     "half bar",
			calls:    50,
			maxCalls: 100,
			width:    16,
			wantLen:  16,
		},
		{
			name:     "empty bar",
			calls:    0,
			maxCalls: 100,
			width:    16,
			wantLen:  16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderUsageBar(tt.calls, tt.maxCalls, tt.width)
			// Remove ANSI codes for length check
			plain := stripANSI(got)
			// Count runes, not bytes (Unicode characters like █ are multi-byte)
			runeCount := len([]rune(plain))
			if runeCount != tt.wantLen {
				t.Errorf("RenderUsageBar() rune count = %d, want %d", runeCount, tt.wantLen)
			}
		})
	}
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// TestRenderRemovalSummary tests removal summary rendering
func TestRenderRemovalSummary(t *testing.T) {
	tests := []struct {
		name    string
		removed []types.MCPServer
		want    []string
	}{
		{
			name:    "no servers removed",
			removed: nil,
			want:    []string{"No servers removed"},
		},
		{
			name: "global and project servers",
			removed: []types.MCPServer{
				{Name: "global-server", Scope: types.ScopeGlobal},
				{Name: "project-server", Scope: types.ScopeProject, ProjectPath: "/path/to/project"},
			},
			want: []string{"global-server", "project-server", "/path/to/project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderRemovalSummary(&buf, tt.removed)
			output := buf.String()

			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

// TestRenderDryRunSummary tests dry run summary rendering
func TestRenderDryRunSummary(t *testing.T) {
	tests := []struct {
		name    string
		servers []types.MCPServer
		want    []string
	}{
		{
			name:    "no servers would be removed",
			servers: nil,
			want:    []string{"No servers would be removed"},
		},
		{
			name: "shows servers with scope",
			servers: []types.MCPServer{
				{Name: "global-server", Scope: types.ScopeGlobal},
				{Name: "project-server", Scope: types.ScopeProject, ProjectPath: "/path/to/project"},
			},
			want: []string{"DRY RUN", "global-server", "global", "project-server", "/path/to/project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderDryRunSummary(&buf, tt.servers)
			output := buf.String()

			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}
