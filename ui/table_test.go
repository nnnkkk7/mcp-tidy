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
					Name:    "context7",
					Type:    types.ServerTypeHTTP,
					URL:     "https://mcp.context7.com/mcp",
					Scope:   types.ScopeGlobal,
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
		name  string
		stats []types.ServerStats
		want  []string
	}{
		{
			name:  "empty stats",
			stats: nil,
			want:  []string{"No usage data"},
		},
		{
			name: "single server stats",
			stats: []types.ServerStats{
				{
					Name:     "context7",
					Calls:    142,
					LastUsed: now.Add(-2 * time.Hour),
				},
			},
			want: []string{"context7", "142", "hours ago"},
		},
		{
			name: "unused server",
			stats: []types.ServerStats{
				{
					Name:     "puppeteer",
					Calls:    0,
					LastUsed: time.Time{},
				},
			},
			want: []string{"puppeteer", "0", "never", "unused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderStatsTable(&buf, tt.stats, 30*24*time.Hour)
			output := buf.String()

			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("RenderStatsTable() output missing %q\nGot:\n%s", want, output)
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
