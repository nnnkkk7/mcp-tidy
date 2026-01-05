package transcript

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nnnkkk7/mcp-tidy/types"
)

func TestExtractServerName(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		wantServer string
		wantTool   string
		wantOK     bool
	}{
		{
			name:       "standard mcp tool name",
			toolName:   "mcp__context7__resolve-library-id",
			wantServer: "context7",
			wantTool:   "resolve-library-id",
			wantOK:     true,
		},
		{
			name:       "another standard mcp tool",
			toolName:   "mcp__serena__find_symbol",
			wantServer: "serena",
			wantTool:   "find_symbol",
			wantOK:     true,
		},
		{
			name:       "three underscores - takes first server segment",
			toolName:   "mcp__my__server__tool__name",
			wantServer: "my",
			wantTool:   "server__tool__name",
			wantOK:     true,
		},
		{
			name:       "non-mcp tool",
			toolName:   "Read",
			wantServer: "",
			wantTool:   "",
			wantOK:     false,
		},
		{
			name:       "mcp prefix but no server",
			toolName:   "mcp__",
			wantServer: "",
			wantTool:   "",
			wantOK:     false,
		},
		{
			name:       "mcp prefix with only server",
			toolName:   "mcp__context7",
			wantServer: "",
			wantTool:   "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServer, gotTool, gotOK := ExtractServerName(tt.toolName)
			if gotOK != tt.wantOK {
				t.Errorf("ExtractServerName() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotServer != tt.wantServer {
				t.Errorf("ExtractServerName() server = %v, want %v", gotServer, tt.wantServer)
			}
			if gotTool != tt.wantTool {
				t.Errorf("ExtractServerName() tool = %v, want %v", gotTool, tt.wantTool)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantCalls []types.ToolCall
		wantErr   bool
	}{
		{
			name: "valid line with mcp tool",
			line: `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_01","name":"mcp__context7__query","input":{}}]},"timestamp":"2025-01-01T10:00:00Z"}`,
			wantCalls: []types.ToolCall{
				{
					ServerName: "context7",
					ToolName:   "query",
					Timestamp:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name:      "line with non-mcp tool",
			line:      `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_01","name":"Read","input":{}}]},"timestamp":"2025-01-01T10:00:00Z"}`,
			wantCalls: nil,
			wantErr:   false,
		},
		{
			name:      "user message (no tool_use)",
			line:      `{"type":"user","message":{"role":"user","content":"hello"},"timestamp":"2025-01-01T10:00:00Z"}`,
			wantCalls: nil,
			wantErr:   false,
		},
		{
			name:      "invalid json",
			line:      `{invalid json}`,
			wantCalls: nil,
			wantErr:   true,
		},
		{
			name:      "empty line",
			line:      "",
			wantCalls: nil,
			wantErr:   false,
		},
		{
			name: "multiple tool calls in one message",
			line: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"mcp__a__t1","input":{}},{"type":"tool_use","name":"mcp__b__t2","input":{}}]},"timestamp":"2025-01-01T10:00:00Z"}`,
			wantCalls: []types.ToolCall{
				{ServerName: "a", ToolName: "t1", Timestamp: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{ServerName: "b", ToolName: "t2", Timestamp: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCalls, err := ParseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantCalls, gotCalls); diff != "" {
				t.Errorf("ParseLine() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantCalls int
		wantErr   bool
	}{
		{
			name:      "parse valid session file",
			filePath:  "../testdata/projects/-Users-xxx-github-proj/session1.jsonl",
			wantCalls: 3, // 2 context7 + 1 serena
			wantErr:   false,
		},
		{
			name:      "parse empty file",
			filePath:  "../testdata/projects/-Users-xxx-github-proj/session2.jsonl",
			wantCalls: 0,
			wantErr:   false,
		},
		{
			name:      "parse file with corrupted line (skip and continue)",
			filePath:  "../testdata/projects/-Users-xxx-github-proj/session3.jsonl",
			wantCalls: 2, // 1 context7 + 1 serena (corrupted line skipped)
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			filePath:  "../testdata/nonexistent.jsonl",
			wantCalls: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls, err := ParseFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(calls) != tt.wantCalls {
				t.Errorf("ParseFile() got %d calls, want %d", len(calls), tt.wantCalls)
			}
		})
	}
}

func TestParseDirectory(t *testing.T) {
	calls, err := ParseDirectory("../testdata/projects")
	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	// Should find 5 total MCP calls (3 from session1 + 2 from session3)
	if len(calls) != 5 {
		t.Errorf("ParseDirectory() got %d calls, want 5", len(calls))
	}

	// Verify servers found
	servers := make(map[string]int)
	for _, call := range calls {
		servers[call.ServerName]++
	}

	wantServers := map[string]int{
		"context7": 3, // 2 from session1 + 1 from session3
		"serena":   2, // 1 from session1 + 1 from session3
	}

	if diff := cmp.Diff(wantServers, servers); diff != "" {
		t.Errorf("ParseDirectory() servers mismatch (-want +got):\n%s", diff)
	}
}

func TestAggregateStats(t *testing.T) {
	calls := []types.ToolCall{
		{ServerName: "context7", ToolName: "query", Timestamp: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ServerName: "context7", ToolName: "resolve", Timestamp: time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)},
		{ServerName: "serena", ToolName: "find", Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
	}

	stats := AggregateStats(calls)

	if len(stats) != 2 {
		t.Fatalf("AggregateStats() got %d servers, want 2", len(stats))
	}

	// Find context7 stats
	var context7Stats, serenaStats types.ServerStats
	for _, s := range stats {
		if s.Name == "context7" {
			context7Stats = s
		}
		if s.Name == "serena" {
			serenaStats = s
		}
	}

	// Check context7
	if context7Stats.Calls != 2 {
		t.Errorf("context7 calls = %d, want 2", context7Stats.Calls)
	}
	if context7Stats.LastUsed != time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC) {
		t.Errorf("context7 lastUsed = %v, want 2025-01-02", context7Stats.LastUsed)
	}

	// Check serena
	if serenaStats.Calls != 1 {
		t.Errorf("serena calls = %d, want 1", serenaStats.Calls)
	}
}

func TestFilterByPeriod(t *testing.T) {
	now := time.Now()
	calls := []types.ToolCall{
		{ServerName: "recent", ToolName: "t1", Timestamp: now.Add(-24 * time.Hour)},
		{ServerName: "old", ToolName: "t2", Timestamp: now.Add(-60 * 24 * time.Hour)},
		{ServerName: "very-old", ToolName: "t3", Timestamp: now.Add(-120 * 24 * time.Hour)},
	}

	tests := []struct {
		name       string
		period     types.Period
		wantCount  int
		wantServer string
	}{
		{
			name:       "7 days - only recent",
			period:     types.Period7Days,
			wantCount:  1,
			wantServer: "recent",
		},
		{
			name:      "30 days - only recent",
			period:    types.Period30Days,
			wantCount: 1,
		},
		{
			name:      "90 days - recent and old",
			period:    types.Period90Days,
			wantCount: 2,
		},
		{
			name:      "all - everything",
			period:    types.PeriodAll,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByPeriod(calls, tt.period)
			if len(filtered) != tt.wantCount {
				t.Errorf("FilterByPeriod() got %d calls, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestDefaultTranscriptPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, ".claude", "projects")
	got := DefaultTranscriptPath()

	if got != want {
		t.Errorf("DefaultTranscriptPath() = %v, want %v", got, want)
	}
}
