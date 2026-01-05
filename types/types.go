// Package types defines common types used across mcp-tidy.
package types

import (
	"fmt"
	"strings"
	"time"
)

// Scope represents the scope of an MCP server configuration.
type Scope int

const (
	// ScopeGlobal indicates a globally configured MCP server.
	ScopeGlobal Scope = iota
	// ScopeProject indicates a project-specific MCP server.
	ScopeProject
)

// String returns the string representation of the scope.
func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "global"
	case ScopeProject:
		return "project"
	default:
		return "unknown"
	}
}

// ServerType represents the type of MCP server connection.
type ServerType int

const (
	// ServerTypeStdio indicates a stdio-based MCP server.
	ServerTypeStdio ServerType = iota
	// ServerTypeHTTP indicates an HTTP-based MCP server.
	ServerTypeHTTP
)

// String returns the string representation of the server type.
func (t ServerType) String() string {
	switch t {
	case ServerTypeStdio:
		return "stdio"
	case ServerTypeHTTP:
		return "http"
	default:
		return "unknown"
	}
}

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Name        string            `json:"name"`
	Type        ServerType        `json:"-"`
	TypeStr     string            `json:"type"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Scope       Scope             `json:"-"`
	ProjectPath string            `json:"-"`
}

// CommandString returns a human-readable representation of the server command.
// For HTTP servers, returns the URL with [http] prefix.
// For stdio servers, returns the command with arguments.
func (s MCPServer) CommandString() string {
	if s.Type == ServerTypeHTTP {
		return fmt.Sprintf("[http] %s", s.URL)
	}

	if len(s.Args) == 0 {
		return s.Command
	}
	return fmt.Sprintf("%s %s", s.Command, strings.Join(s.Args, " "))
}

// ScopeString returns the scope as a display string.
// For global scope, returns "global".
// For project scope, returns the project path.
func (s MCPServer) ScopeString() string {
	if s.Scope == ScopeGlobal {
		return "global"
	}
	return s.ProjectPath
}

// ServerStats holds usage statistics for an MCP server.
type ServerStats struct {
	Name     string
	Calls    int
	LastUsed time.Time
	Tools    map[string]int // tool name -> call count
}

// IsUnused returns true if the server hasn't been used within the given period.
func (s ServerStats) IsUnused(period time.Duration) bool {
	if s.LastUsed.IsZero() {
		return true
	}
	return time.Since(s.LastUsed) > period
}

// LastUsedString returns a human-readable representation of when the server was last used.
func (s ServerStats) LastUsedString() string {
	if s.LastUsed.IsZero() {
		return "never"
	}

	duration := time.Since(s.LastUsed)
	hours := int(duration.Hours())
	days := hours / 24

	if days > 0 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	if hours > 0 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	minutes := int(duration.Minutes())
	if minutes > 0 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	return "just now"
}

// ToolCall represents a single MCP tool invocation extracted from logs.
type ToolCall struct {
	ServerName string
	ToolName   string
	Timestamp  time.Time
}

// Period represents a time period for filtering stats.
type Period int

const (
	// Period7Days represents a 7-day period.
	Period7Days Period = iota
	// Period30Days represents a 30-day period (default).
	Period30Days
	// Period90Days represents a 90-day period.
	Period90Days
	// PeriodAll represents all time.
	PeriodAll
)

// Duration returns the time.Duration for the period.
func (p Period) Duration() time.Duration {
	switch p {
	case Period7Days:
		return 7 * 24 * time.Hour
	case Period30Days:
		return 30 * 24 * time.Hour
	case Period90Days:
		return 90 * 24 * time.Hour
	case PeriodAll:
		return 0 // special case: no filtering
	default:
		return 30 * 24 * time.Hour
	}
}

// ParsePeriod parses a period string (7d, 30d, 90d, all) into a Period.
func ParsePeriod(s string) Period {
	switch strings.ToLower(s) {
	case "7d":
		return Period7Days
	case "30d":
		return Period30Days
	case "90d":
		return Period90Days
	case "all":
		return PeriodAll
	default:
		return Period30Days
	}
}
