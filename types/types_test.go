package types

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestScope_String(t *testing.T) {
	tests := []struct {
		name  string
		scope Scope
		want  string
	}{
		{
			name:  "global scope",
			scope: ScopeGlobal,
			want:  "global",
		},
		{
			name:  "project scope",
			scope: ScopeProject,
			want:  "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scope.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Scope.String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServerType_String(t *testing.T) {
	tests := []struct {
		name       string
		serverType ServerType
		want       string
	}{
		{
			name:       "stdio type",
			serverType: ServerTypeStdio,
			want:       "stdio",
		},
		{
			name:       "http type",
			serverType: ServerTypeHTTP,
			want:       "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.serverType.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ServerType.String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMCPServer_CommandString(t *testing.T) {
	tests := []struct {
		name   string
		server MCPServer
		want   string
	}{
		{
			name: "http server returns url",
			server: MCPServer{
				Name: "context7",
				Type: ServerTypeHTTP,
				URL:  "https://mcp.context7.com/mcp",
			},
			want: "[http] https://mcp.context7.com/mcp",
		},
		{
			name: "stdio server returns command with args",
			server: MCPServer{
				Name:    "serena",
				Type:    ServerTypeStdio,
				Command: "uvx",
				Args:    []string{"--from", "git+https://github.com/oraios/serena", "serena"},
			},
			want: "uvx --from git+https://github.com/oraios/serena serena",
		},
		{
			name: "stdio server with no args",
			server: MCPServer{
				Name:    "simple",
				Type:    ServerTypeStdio,
				Command: "npx",
			},
			want: "npx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.server.CommandString()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MCPServer.CommandString() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMCPServer_ScopeString(t *testing.T) {
	tests := []struct {
		name   string
		server MCPServer
		want   string
	}{
		{
			name: "global scope",
			server: MCPServer{
				Name:  "context7",
				Scope: ScopeGlobal,
			},
			want: "global",
		},
		{
			name: "project scope with path",
			server: MCPServer{
				Name:        "serena",
				Scope:       ScopeProject,
				ProjectPath: "/Users/xxx/github/my-project",
			},
			want: "/Users/xxx/github/my-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.server.ScopeString()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MCPServer.ScopeString() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServerStats_IsUnused(t *testing.T) {
	now := time.Now()
	twentyNineDaysAgo := now.AddDate(0, 0, -29)
	thirtyOneDaysAgo := now.AddDate(0, 0, -31)

	tests := []struct {
		name   string
		stats  ServerStats
		period time.Duration
		want   bool
	}{
		{
			name: "never used is unused",
			stats: ServerStats{
				Name:     "puppeteer",
				Calls:    0,
				LastUsed: time.Time{}, // zero time = never used
			},
			period: 30 * 24 * time.Hour,
			want:   true,
		},
		{
			name: "used within period is not unused",
			stats: ServerStats{
				Name:     "context7",
				Calls:    10,
				LastUsed: twentyNineDaysAgo,
			},
			period: 30 * 24 * time.Hour,
			want:   false,
		},
		{
			name: "used outside period is unused",
			stats: ServerStats{
				Name:     "old-server",
				Calls:    5,
				LastUsed: thirtyOneDaysAgo,
			},
			period: 30 * 24 * time.Hour,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.IsUnused(tt.period)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ServerStats.IsUnused() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServerStats_LastUsedString(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		stats ServerStats
		want  string
	}{
		{
			name: "never used",
			stats: ServerStats{
				Name:     "puppeteer",
				LastUsed: time.Time{},
			},
			want: "never",
		},
		{
			name: "used recently",
			stats: ServerStats{
				Name:     "context7",
				LastUsed: now.Add(-2 * time.Hour),
			},
			want: "2 hours ago",
		},
		{
			name: "used 1 day ago",
			stats: ServerStats{
				Name:     "serena",
				LastUsed: now.Add(-24 * time.Hour),
			},
			want: "1 day ago",
		},
		{
			name: "used multiple days ago",
			stats: ServerStats{
				Name:     "old",
				LastUsed: now.Add(-72 * time.Hour),
			},
			want: "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.LastUsedString()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ServerStats.LastUsedString() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
