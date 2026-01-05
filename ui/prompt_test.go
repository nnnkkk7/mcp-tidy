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
	servers := []types.MCPServer{
		{Name: "context7", Scope: types.ScopeGlobal},
		{Name: "serena", Scope: types.ScopeProject, ProjectPath: "/path"},
		{Name: "puppeteer", Scope: types.ScopeGlobal},
	}

	stats := map[string]types.ServerStats{
		"context7":  {Name: "context7", Calls: 100},
		"puppeteer": {Name: "puppeteer", Calls: 0},
	}

	tests := []struct {
		name    string
		input   string
		want    []int
		wantLen int
	}{
		{
			name:    "select single",
			input:   "1\n",
			want:    []int{0},
			wantLen: 1,
		},
		{
			name:    "select multiple",
			input:   "1 3\n",
			want:    []int{0, 2},
			wantLen: 2,
		},
		{
			name:    "select all",
			input:   "all\n",
			wantLen: 3,
		},
		{
			name:    "empty input",
			input:   "\n",
			wantLen: 0,
		},
		{
			name:    "invalid number ignored",
			input:   "1 99 2\n",
			want:    []int{0, 1},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}

			got := SelectServersPromptWithReader(r, w, servers, stats)

			if len(got) != tt.wantLen {
				t.Errorf("SelectServersPromptWithReader() returned %d items, want %d", len(got), tt.wantLen)
			}

			if tt.want != nil {
				for i, idx := range tt.want {
					if i < len(got) && got[i] != idx {
						t.Errorf("SelectServersPromptWithReader()[%d] = %d, want %d", i, got[i], idx)
					}
				}
			}
		})
	}
}
