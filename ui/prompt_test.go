package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nnnkkk7/mcp-tidy/types"
)

func TestConfirmPromptWithReader(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		want       bool
	}{
		{
			name:       "yes input",
			input:      "y\n",
			defaultYes: false,
			want:       true,
		},
		{
			name:       "YES input",
			input:      "YES\n",
			defaultYes: false,
			want:       true,
		},
		{
			name:       "no input",
			input:      "n\n",
			defaultYes: true,
			want:       false,
		},
		{
			name:       "empty input with default no",
			input:      "\n",
			defaultYes: false,
			want:       false,
		},
		{
			name:       "empty input with default yes",
			input:      "\n",
			defaultYes: true,
			want:       true,
		},
		{
			name:       "invalid input defaults to no",
			input:      "maybe\n",
			defaultYes: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}

			got := ConfirmPromptWithReader(r, w, "Remove?", tt.defaultYes)
			if got != tt.want {
				t.Errorf("ConfirmPromptWithReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectServersPromptWithReader(t *testing.T) {
	tests := []struct {
		name        string
		servers     []types.MCPServer
		stats       map[string]types.ServerStats
		input       string
		wantIndices []int
		wantLen     int
		wantOutput  []string       // substrings that should be present in output
		notWant     []string       // substrings that should NOT be present
		minCount    map[string]int // minimum occurrence count for specific strings
	}{
		// Selection tests
		{
			name: "select single",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path"},
				{Name: "puppeteer", Scope: types.ScopeGlobal},
			},
			stats: map[string]types.ServerStats{
				"context7":  {Name: "context7", Calls: 100},
				"puppeteer": {Name: "puppeteer", Calls: 0},
			},
			input:       "1\n",
			wantIndices: []int{0},
			wantLen:     1,
		},
		{
			name: "select multiple",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path"},
				{Name: "puppeteer", Scope: types.ScopeGlobal},
			},
			stats: map[string]types.ServerStats{
				"context7":  {Name: "context7", Calls: 100},
				"puppeteer": {Name: "puppeteer", Calls: 0},
			},
			input:       "1 3\n",
			wantIndices: []int{0, 2},
			wantLen:     2,
		},
		{
			name: "select all",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path"},
				{Name: "puppeteer", Scope: types.ScopeGlobal},
			},
			stats:   map[string]types.ServerStats{},
			input:   "all\n",
			wantLen: 3,
		},
		{
			name: "empty input returns empty",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
			},
			stats:   map[string]types.ServerStats{},
			input:   "\n",
			wantLen: 0,
		},
		{
			name: "invalid number ignored",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path"},
				{Name: "puppeteer", Scope: types.ScopeGlobal},
			},
			stats:       map[string]types.ServerStats{},
			input:       "1 99 2\n",
			wantIndices: []int{0, 1},
			wantLen:     2,
		},
		{
			name:    "empty servers returns nil",
			servers: []types.MCPServer{},
			stats:   map[string]types.ServerStats{},
			input:   "1\n",
			wantLen: 0,
		},
		// Output format tests - scope display
		{
			name: "shows global and project scope",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/Users/test/project"},
			},
			stats: map[string]types.ServerStats{
				"context7": {Name: "context7", Calls: 100},
				"serena":   {Name: "serena", Calls: 50},
			},
			input:      "1\n",
			wantLen:    1,
			wantOutput: []string{"[global]", "/Users/test/project"},
		},
		{
			name: "truncates long project path",
			servers: []types.MCPServer{
				{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/Users/username/Documents/projects/very-long-project-name-that-exceeds-limit"},
			},
			stats: map[string]types.ServerStats{
				"serena": {Name: "serena", Calls: 50},
			},
			input:      "1\n",
			wantLen:    1,
			wantOutput: []string{"..."},
			notWant:    []string{"/Users/username/Documents/projects/very-long-project-name-that-exceeds-limit"},
		},
		// Output format tests - usage info
		{
			name: "shows usage info for servers",
			servers: []types.MCPServer{
				{Name: "used-server", Scope: types.ScopeGlobal},
				{Name: "unused-server", Scope: types.ScopeGlobal},
				{Name: "no-stats-server", Scope: types.ScopeGlobal},
			},
			stats: map[string]types.ServerStats{
				"used-server":   {Name: "used-server", Calls: 100},
				"unused-server": {Name: "unused-server", Calls: 0},
			},
			input:      "1\n",
			wantLen:    1,
			wantOutput: []string{"100 calls", "0 calls", "unused", "no usage data"},
		},
		// Output format tests - same name different scope
		{
			name: "distinguishes same name servers by scope",
			servers: []types.MCPServer{
				{Name: "context7", Scope: types.ScopeGlobal},
				{Name: "context7", Scope: types.ScopeProject, ProjectPath: "/project-a"},
				{Name: "context7", Scope: types.ScopeProject, ProjectPath: "/project-b"},
			},
			stats: map[string]types.ServerStats{
				"context7": {Name: "context7", Calls: 100},
			},
			input:      "1\n",
			wantLen:    1,
			wantOutput: []string{"[global]", "/project-a", "/project-b"},
			minCount:   map[string]int{"context7": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}

			got := SelectServersPromptWithReader(r, w, tt.servers, tt.stats)
			output := w.String()

			// Check return value length
			if len(got) != tt.wantLen {
				t.Errorf("returned %d items, want %d", len(got), tt.wantLen)
			}

			// Check specific indices
			if tt.wantIndices != nil {
				for i, idx := range tt.wantIndices {
					if i < len(got) && got[i] != idx {
						t.Errorf("index[%d] = %d, want %d", i, got[i], idx)
					}
				}
			}

			// Check output contains expected strings
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}

			// Check output does not contain unwanted strings
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
		})
	}
}
