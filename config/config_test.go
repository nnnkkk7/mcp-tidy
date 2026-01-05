package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nnnkkk7/mcp-tidy/types"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		wantServers []types.MCPServer
		wantErr     bool
	}{
		{
			name:       "load full config with global and project servers",
			configPath: "../testdata/claude.json",
			wantServers: []types.MCPServer{
				{
					Name:    "context7",
					Type:    types.ServerTypeHTTP,
					TypeStr: "http",
					URL:     "https://mcp.context7.com/mcp",
					Scope:   types.ScopeGlobal,
				},
				{
					Name:    "puppeteer",
					Type:    types.ServerTypeStdio,
					TypeStr: "stdio",
					Command: "npx",
					Args:    []string{"-y", "@anthropic/server-puppeteer"},
					Scope:   types.ScopeGlobal,
				},
				{
					Name:        "serena",
					Type:        types.ServerTypeStdio,
					TypeStr:     "stdio",
					Command:     "uvx",
					Args:        []string{"--from", "git+https://github.com/oraios/serena", "serena", "start-mcp-server"},
					Env:         map[string]string{},
					Scope:       types.ScopeProject,
					ProjectPath: "/Users/xxx/github/my-project",
				},
			},
			wantErr: false,
		},
		{
			name:        "load empty config",
			configPath:  "../testdata/claude_empty.json",
			wantServers: nil,
			wantErr:     false,
		},
		{
			name:        "file not exists returns empty config",
			configPath:  "../testdata/nonexistent.json",
			wantServers: nil,
			wantErr:     false,
		},
		{
			name:       "load global only config",
			configPath: "../testdata/claude_global_only.json",
			wantServers: []types.MCPServer{
				{
					Name:    "context7",
					Type:    types.ServerTypeHTTP,
					TypeStr: "http",
					URL:     "https://mcp.context7.com/mcp",
					Scope:   types.ScopeGlobal,
				},
			},
			wantErr: false,
		},
		{
			name:       "load project only config",
			configPath: "../testdata/claude_project_only.json",
			wantServers: []types.MCPServer{
				{
					Name:        "serena",
					Type:        types.ServerTypeStdio,
					TypeStr:     "stdio",
					Command:     "uvx",
					Args:        []string{"--from", "git+https://github.com/oraios/serena", "serena"},
					Scope:       types.ScopeProject,
					ProjectPath: "/Users/xxx/github/my-project",
				},
			},
			wantErr: false,
		},
		{
			name:        "invalid json returns error",
			configPath:  "../testdata/claude_invalid.json",
			wantServers: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			got := cfg.Servers()

			// Sort servers for consistent comparison
			sortOpts := cmp.Transformer("sortServers", func(servers []types.MCPServer) []types.MCPServer {
				sorted := make([]types.MCPServer, len(servers))
				copy(sorted, servers)
				// Simple sort by name
				for i := 0; i < len(sorted)-1; i++ {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[i].Name > sorted[j].Name {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}
				return sorted
			})

			if diff := cmp.Diff(tt.wantServers, got, sortOpts); diff != "" {
				t.Errorf("Load() servers mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfig_GetServer(t *testing.T) {
	cfg, err := Load("../testdata/claude.json")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	tests := []struct {
		name       string
		serverName string
		wantFound  bool
		wantType   types.ServerType
	}{
		{
			name:       "find global http server",
			serverName: "context7",
			wantFound:  true,
			wantType:   types.ServerTypeHTTP,
		},
		{
			name:       "find global stdio server",
			serverName: "puppeteer",
			wantFound:  true,
			wantType:   types.ServerTypeStdio,
		},
		{
			name:       "find project server",
			serverName: "serena",
			wantFound:  true,
			wantType:   types.ServerTypeStdio,
		},
		{
			name:       "not found",
			serverName: "nonexistent",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, found := cfg.GetServer(tt.serverName)
			if found != tt.wantFound {
				t.Errorf("GetServer() found = %v, want %v", found, tt.wantFound)
				return
			}
			if tt.wantFound && server.Type != tt.wantType {
				t.Errorf("GetServer() type = %v, want %v", server.Type, tt.wantType)
			}
		})
	}
}

func TestDefaultConfigPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	want := filepath.Join(homeDir, ".claude.json")
	got := DefaultConfigPath()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DefaultConfigPath() mismatch (-want +got):\n%s", diff)
	}
}
