package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nnnkkk7/mcp-tidy/types"
)

func TestBackup(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "mcp-tidy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test config file
	configPath := filepath.Join(tmpDir, "claude.json")
	content := `{"mcpServers": {"test": {"type": "http", "url": "https://test.com"}}}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "creates backup file",
			configPath: configPath,
			wantErr:    false,
		},
		{
			name:       "non-existent file returns error",
			configPath: filepath.Join(tmpDir, "nonexistent.json"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupPath, err := Backup(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Backup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Verify backup file exists
			if _, err := os.Stat(backupPath); err != nil {
				t.Errorf("backup file does not exist: %s", backupPath)
			}

			// Verify backup content matches original
			backupContent, err := os.ReadFile(backupPath)
			if err != nil {
				t.Fatalf("failed to read backup: %v", err)
			}
			if diff := cmp.Diff(content, string(backupContent)); diff != "" {
				t.Errorf("backup content mismatch (-want +got):\n%s", diff)
			}

			// Verify backup filename format
			if !strings.Contains(backupPath, ".backup.") {
				t.Errorf("backup path should contain '.backup.': %s", backupPath)
			}
		})
	}
}

func TestBackup_NoOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-tidy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "claude.json")
	if err := os.WriteFile(configPath, []byte(`{"test": true}`), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Create first backup
	backup1, err := Backup(configPath)
	if err != nil {
		t.Fatalf("first backup failed: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(time.Second)

	// Change the content
	if err := os.WriteFile(configPath, []byte(`{"test": false}`), 0o644); err != nil {
		t.Fatalf("failed to update test config: %v", err)
	}

	// Create second backup
	backup2, err := Backup(configPath)
	if err != nil {
		t.Fatalf("second backup failed: %v", err)
	}

	// Both backups should exist and be different
	if backup1 == backup2 {
		t.Errorf("backups should have different names: %s vs %s", backup1, backup2)
	}

	content1, _ := os.ReadFile(backup1)
	content2, _ := os.ReadFile(backup2)

	if bytes.Equal(content1, content2) {
		t.Error("backup contents should be different")
	}
}

func TestRemoveServer(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  string
		serverToRemove string
		scope          types.Scope
		projectPath    string
		wantConfig     map[string]interface{}
		wantErr        bool
	}{
		{
			name: "remove global server",
			initialConfig: `{
				"mcpServers": {
					"context7": {"type": "http", "url": "https://test.com"},
					"puppeteer": {"type": "stdio", "command": "npx"}
				}
			}`,
			serverToRemove: "context7",
			scope:          types.ScopeGlobal,
			wantConfig: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"puppeteer": map[string]interface{}{"type": "stdio", "command": "npx"},
				},
			},
			wantErr: false,
		},
		{
			name: "remove project server",
			initialConfig: `{
				"projects": {
					"/path/to/project": {
						"mcpServers": {
							"serena": {"type": "stdio", "command": "uvx"},
							"other": {"type": "stdio", "command": "other"}
						}
					}
				}
			}`,
			serverToRemove: "serena",
			scope:          types.ScopeProject,
			projectPath:    "/path/to/project",
			wantConfig: map[string]interface{}{
				"projects": map[string]interface{}{
					"/path/to/project": map[string]interface{}{
						"mcpServers": map[string]interface{}{
							"other": map[string]interface{}{"type": "stdio", "command": "other"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "remove last global server",
			initialConfig: `{
				"mcpServers": {
					"only": {"type": "http", "url": "https://test.com"}
				}
			}`,
			serverToRemove: "only",
			scope:          types.ScopeGlobal,
			wantConfig: map[string]interface{}{
				"mcpServers": map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "remove non-existent server (warning only)",
			initialConfig: `{
				"mcpServers": {
					"existing": {"type": "http", "url": "https://test.com"}
				}
			}`,
			serverToRemove: "nonexistent",
			scope:          types.ScopeGlobal,
			wantConfig: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"existing": map[string]interface{}{"type": "http", "url": "https://test.com"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "mcp-tidy-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			configPath := filepath.Join(tmpDir, "claude.json")
			if err := os.WriteFile(configPath, []byte(tt.initialConfig), 0o644); err != nil {
				t.Fatalf("failed to write initial config: %v", err)
			}

			server := &types.MCPServer{
				Name:        tt.serverToRemove,
				Scope:       tt.scope,
				ProjectPath: tt.projectPath,
			}

			err = RemoveServer(configPath, server)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Read and parse the result
			result, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("failed to read result: %v", err)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(result, &got); err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			if diff := cmp.Diff(tt.wantConfig, got); diff != "" {
				t.Errorf("RemoveServer() result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-tidy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "claude.json")
	content := []byte(`{"test": "atomic write"}`)

	if err := atomicWrite(configPath, content); err != nil {
		t.Fatalf("atomicWrite() failed: %v", err)
	}

	// Verify file was written correctly
	result, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	if diff := cmp.Diff(string(content), string(result)); diff != "" {
		t.Errorf("atomicWrite() content mismatch (-want +got):\n%s", diff)
	}

	// Verify no temp files remain
	files, _ := os.ReadDir(tmpDir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			t.Errorf("temp file should be removed: %s", f.Name())
		}
	}
}
