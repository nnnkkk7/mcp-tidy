// Package transcript handles parsing Claude Code transcript logs.
package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nnnkkk7/mcp-tidy/types"
)

// logEntry represents a single entry in the JSONL log file.
type logEntry struct {
	Type      string    `json:"type"`
	Message   message   `json:"message"`
	Timestamp string    `json:"timestamp"`
	UUID      string    `json:"uuid"`
}

// message represents the message field in a log entry.
type message struct {
	Role       string          `json:"role"`
	ContentRaw json.RawMessage `json:"content"`
	Content    []content       `json:"-"`
}

// content represents a single content item in a message.
type content struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// DefaultTranscriptPath returns the default path to Claude transcript logs.
func DefaultTranscriptPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude", "projects")
}

// ExtractServerName extracts the server name and tool name from an MCP tool name.
// MCP tool names follow the pattern: mcp__{server}__{tool}
// Returns (serverName, toolName, ok).
func ExtractServerName(toolName string) (string, string, bool) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return "", "", false
	}

	// Remove "mcp__" prefix
	rest := strings.TrimPrefix(toolName, "mcp__")

	// Find the first "__" to split server from tool
	idx := strings.Index(rest, "__")
	if idx == -1 || idx == 0 {
		return "", "", false
	}

	server := rest[:idx]
	tool := rest[idx+2:]

	if server == "" || tool == "" {
		return "", "", false
	}

	return server, tool, true
}

// ParseLine parses a single JSONL line and extracts MCP tool calls.
func ParseLine(line string) ([]types.ToolCall, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	var entry logEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return nil, fmt.Errorf("failed to parse line: %w", err)
	}

	// Parse content - it could be an array or a string
	// Only arrays contain tool_use entries
	if len(entry.Message.ContentRaw) > 0 && entry.Message.ContentRaw[0] == '[' {
		if err := json.Unmarshal(entry.Message.ContentRaw, &entry.Message.Content); err != nil {
			// If parsing fails, it's likely a different format - skip silently
			return nil, nil
		}
	} else {
		// Content is a string or other type, no tool_use entries
		return nil, nil
	}

	// Parse timestamp
	var timestamp time.Time
	if entry.Timestamp != "" {
		var err error
		timestamp, err = time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			// Try RFC3339Nano
			timestamp, err = time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err != nil {
				timestamp = time.Time{}
			}
		}
	}

	var calls []types.ToolCall

	// Extract tool_use entries from message.content
	for _, c := range entry.Message.Content {
		if c.Type != "tool_use" {
			continue
		}

		serverName, toolName, ok := ExtractServerName(c.Name)
		if !ok {
			continue
		}

		calls = append(calls, types.ToolCall{
			ServerName: serverName,
			ToolName:   toolName,
			Timestamp:  timestamp,
		})
	}

	return calls, nil
}

// ParseFile parses a JSONL file and extracts all MCP tool calls.
// Corrupted lines are skipped with a warning.
func ParseFile(filePath string) ([]types.ToolCall, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Check if file is empty
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.Size() == 0 {
		return nil, nil
	}

	var allCalls []types.ToolCall
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max line size

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		calls, err := ParseLine(line)
		if err != nil {
			// Skip corrupted lines (warning would be printed in production)
			continue
		}

		allCalls = append(allCalls, calls...)
	}

	if err := scanner.Err(); err != nil {
		return allCalls, fmt.Errorf("error reading file: %w", err)
	}

	return allCalls, nil
}

// ParseDirectory parses all JSONL files in a directory and its subdirectories.
func ParseDirectory(dirPath string) ([]types.ToolCall, error) {
	var allCalls []types.ToolCall

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .jsonl files
		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		// Skip empty files
		if info.Size() == 0 {
			return nil
		}

		calls, err := ParseFile(path)
		if err != nil {
			// Log warning but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		allCalls = append(allCalls, calls...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allCalls, nil
}

// AggregateStats aggregates tool calls into per-server statistics.
func AggregateStats(calls []types.ToolCall) []types.ServerStats {
	statsMap := make(map[string]*types.ServerStats)

	for _, call := range calls {
		stats, ok := statsMap[call.ServerName]
		if !ok {
			stats = &types.ServerStats{
				Name:  call.ServerName,
				Tools: make(map[string]int),
			}
			statsMap[call.ServerName] = stats
		}

		stats.Calls++
		stats.Tools[call.ToolName]++

		// Update last used time if this call is more recent
		if call.Timestamp.After(stats.LastUsed) {
			stats.LastUsed = call.Timestamp
		}
	}

	// Convert map to slice
	result := make([]types.ServerStats, 0, len(statsMap))
	for _, stats := range statsMap {
		result = append(result, *stats)
	}

	return result
}

// FilterByPeriod filters tool calls by time period.
func FilterByPeriod(calls []types.ToolCall, period types.Period) []types.ToolCall {
	if period == types.PeriodAll {
		return calls
	}

	duration := period.Duration()
	cutoff := time.Now().Add(-duration)

	var filtered []types.ToolCall
	for _, call := range calls {
		if call.Timestamp.After(cutoff) {
			filtered = append(filtered, call)
		}
	}

	return filtered
}

// GetStats parses all transcripts and returns aggregated statistics.
func GetStats(transcriptPath string, period types.Period) ([]types.ServerStats, error) {
	calls, err := ParseDirectory(transcriptPath)
	if err != nil {
		return nil, err
	}

	filtered := FilterByPeriod(calls, period)
	stats := AggregateStats(filtered)

	return stats, nil
}
