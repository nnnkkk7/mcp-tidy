// Package config handles reading and writing Claude Code MCP configuration.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/nnnkkk7/mcp-tidy/types"
)

// rawServerConfig represents the raw JSON structure of an MCP server.
type rawServerConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// rawProjectConfig represents a project's configuration.
type rawProjectConfig struct {
	MCPServers map[string]rawServerConfig `json:"mcpServers,omitempty"`
}

// rawConfig represents the raw JSON structure of ~/.claude.json.
type rawConfig struct {
	MCPServers map[string]rawServerConfig  `json:"mcpServers,omitempty"`
	Projects   map[string]rawProjectConfig `json:"projects,omitempty"`
}

// Config holds the parsed MCP server configuration.
type Config struct {
	path       string
	raw        rawConfig
	servers    []types.MCPServer
	serverMap  map[string]types.MCPServer
	rawContent []byte
}

// DefaultConfigPath returns the default path to the Claude configuration file.
func DefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude.json")
}

// Load reads and parses the Claude configuration file.
// If the file does not exist, returns an empty config (not an error).
func Load(path string) (*Config, error) {
	cfg := &Config{
		path:      path,
		serverMap: make(map[string]types.MCPServer),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}

	cfg.rawContent = data

	if err := json.Unmarshal(data, &cfg.raw); err != nil {
		return nil, err
	}

	cfg.parseServers()
	return cfg, nil
}

// parseServers converts raw server configs into typed MCPServer structs.
func (c *Config) parseServers() {
	// Parse global servers
	for name, raw := range c.raw.MCPServers {
		server := c.parseServer(name, &raw, types.ScopeGlobal, "")
		c.servers = append(c.servers, server)
		c.serverMap[name] = server
	}

	// Parse project-specific servers
	for projectPath, project := range c.raw.Projects {
		for name, raw := range project.MCPServers {
			server := c.parseServer(name, &raw, types.ScopeProject, projectPath)
			c.servers = append(c.servers, server)
			c.serverMap[name] = server
		}
	}
}

// parseServer converts a raw server config into a typed MCPServer.
func (c *Config) parseServer(name string, raw *rawServerConfig, scope types.Scope, projectPath string) types.MCPServer {
	serverType := types.ServerTypeStdio
	if raw.Type == "http" {
		serverType = types.ServerTypeHTTP
	}

	return types.MCPServer{
		Name:        name,
		Type:        serverType,
		TypeStr:     raw.Type,
		Command:     raw.Command,
		Args:        raw.Args,
		Env:         raw.Env,
		URL:         raw.URL,
		Headers:     raw.Headers,
		Scope:       scope,
		ProjectPath: projectPath,
	}
}

// Servers returns all configured MCP servers.
func (c *Config) Servers() []types.MCPServer {
	return c.servers
}

// GetServer returns a server by name.
func (c *Config) GetServer(name string) (types.MCPServer, bool) {
	server, ok := c.serverMap[name]
	return server, ok
}

// GlobalServers returns only globally configured servers.
func (c *Config) GlobalServers() []types.MCPServer {
	var result []types.MCPServer
	for i := range c.servers {
		if c.servers[i].Scope == types.ScopeGlobal {
			result = append(result, c.servers[i])
		}
	}
	return result
}

// ProjectServers returns servers for a specific project path.
func (c *Config) ProjectServers(projectPath string) []types.MCPServer {
	var result []types.MCPServer
	for i := range c.servers {
		if c.servers[i].Scope == types.ScopeProject && c.servers[i].ProjectPath == projectPath {
			result = append(result, c.servers[i])
		}
	}
	return result
}

// Path returns the config file path.
func (c *Config) Path() string {
	return c.path
}

// RawContent returns the original JSON content.
func (c *Config) RawContent() []byte {
	return c.rawContent
}
