package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nnnkkk7/mcp-tidy/types"
)

// Backup creates a backup of the config file.
// Returns the path to the backup file.
// Backup filename format: {original}.backup.{YYYYMMDD-HHMMSS}
func Backup(configPath string) (string, error) {
	// Read original content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config for backup: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", configPath, timestamp)

	// Write backup file
	if err := os.WriteFile(backupPath, content, 0o644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// RemoveServer removes a server from the config file.
// For global servers, removes from mcpServers.
// For project servers, removes from projects.{path}.mcpServers.
func RemoveServer(configPath string, server *types.MCPServer) error {
	// Read and parse the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Remove the server based on scope
	if server.Scope == types.ScopeGlobal {
		if mcpServers, ok := raw["mcpServers"].(map[string]interface{}); ok {
			delete(mcpServers, server.Name)
		}
	} else {
		// Project scope
		if projects, ok := raw["projects"].(map[string]interface{}); ok {
			if project, ok := projects[server.ProjectPath].(map[string]interface{}); ok {
				if mcpServers, ok := project["mcpServers"].(map[string]interface{}); ok {
					delete(mcpServers, server.Name)
				}
			}
		}
	}

	// Marshal back to JSON with indentation
	newContent, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically
	if err := atomicWrite(configPath, newContent); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// RemoveServers removes multiple servers from the config file.
// Creates a single backup before removing all servers.
func RemoveServers(configPath string, servers []types.MCPServer) error {
	if len(servers) == 0 {
		return nil
	}

	// Create backup first
	backupPath, err := Backup(configPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("Backup created: %s\n", backupPath)

	// Read and parse the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Remove each server
	for i := range servers {
		if servers[i].Scope == types.ScopeGlobal {
			if mcpServers, ok := raw["mcpServers"].(map[string]interface{}); ok {
				delete(mcpServers, servers[i].Name)
			}
		} else {
			if projects, ok := raw["projects"].(map[string]interface{}); ok {
				if project, ok := projects[servers[i].ProjectPath].(map[string]interface{}); ok {
					if mcpServers, ok := project["mcpServers"].(map[string]interface{}); ok {
						delete(mcpServers, servers[i].Name)
					}
				}
			}
		}
	}

	// Marshal back to JSON with indentation
	newContent, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically
	if err := atomicWrite(configPath, newContent); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// atomicWrite writes content to a file atomically using a temp file and rename.
// This prevents data loss if Claude Code reads the file during write.
func atomicWrite(path string, content []byte) error {
	// Create temp file in the same directory
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "mcp-tidy-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write content to temp file
	if _, err := tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Close temp file before rename
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to target path (atomic on most filesystems)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
