# mcp-tidy

[![CI](https://github.com/nnnkkk7/mcp-tidy/actions/workflows/ci.yaml/badge.svg)](https://github.com/nnnkkk7/mcp-tidy/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/mcp-tidy)](https://goreportcard.com/report/github.com/nnnkkk7/mcp-tidy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

![mcp-tidy demo](assets/demo.gif)

## Why mcp-tidy?

### The Problem: MCP Server Bloat

As you use Claude Code, MCP servers accumulate in your `~/.claude.json`. This causes real problems:

| Issue | Impact |
|-------|--------|
| **Context window consumption** | Claude Code can load all tool descriptions (built-in + MCP) into context immediately after the first message, creating ~10k–20k token overhead even for simple queries (e.g., ~11.6k without MCP, ~15k with 4 MCP servers, ~20k+ with 7+). [^1] |
| **Degraded tool selection** | More available tools increases the chance of wrong tool selection / incorrect parameters, especially when tools have similar names. [^2] |
| **Higher latency** | As MCP usage scales, common patterns can increase agent cost and latency (e.g., tool definitions overloading context; intermediate tool results consuming additional tokens). [^3] |

[^1]: [Built-in tools + MCP descriptions load on first message causing 10-20k token overhead](https://github.com/anthropics/claude-code/issues/3406)
[^2]: [Introducing advanced tool use on the Claude Developer Console](https://www.anthropic.com/engineering/advanced-tool-use) - Tool selection failures increase with more available tools (especially with similar names)
[^3]: [Code execution with MCP: building more efficient AI agents](https://www.anthropic.com/engineering/code-execution-with-mcp)

### The Solution

**mcp-tidy** helps you reclaim context and improve Claude Code performance:

- **See** what MCP servers are configured (global and per-project)
- **Understand** which ones you actually use (with call statistics)
- **Clean up** unused servers safely (with automatic backups)

> **Note**: Currently supports **Claude Code** only. Other MCP clients (Claude Desktop, etc.) are not yet supported.


## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Limitations](#limitations)
- [Contributing](#contributing)
- [License](#license)

## Features

| Command | Description |
|---------|-------------|
| `mcp-tidy list` | Display all configured MCP servers (global + project-scoped) |
| `mcp-tidy stats` | Show usage statistics with visual usage bars |
| `mcp-tidy remove` | Interactively remove unused servers with backup |

## Quick Start

```bash
# Install
brew install nnnkkk7/tap/mcp-tidy

# See your MCP servers
mcp-tidy list

# Check which ones you actually use
mcp-tidy stats

# Clean up unused servers
mcp-tidy remove --unused
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install nnnkkk7/tap/mcp-tidy
```

### Go Install

```bash
go install github.com/nnnkkk7/mcp-tidy/cmd/mcp-tidy@latest
```

### From Source

```bash
git clone https://github.com/nnnkkk7/mcp-tidy.git
cd mcp-tidy
make install
```

### Download Binary

Download the latest binary from [GitHub Releases](https://github.com/nnnkkk7/mcp-tidy/releases).

## Usage

### List MCP Servers

```bash
mcp-tidy list
```

Output:

```
MCP Servers (3 configured)

  NAME       TYPE    SCOPE                         COMMAND
  context7   [http]  global                        https://mcp.context7.com/mcp
  serena     [stdio] /Users/xxx/github/project     uvx --from git+https://...
  puppeteer  [stdio] global                        npx @anthropic/puppeteer
```

### View Usage Statistics

```bash
mcp-tidy stats
```

Output (grouped by scope):

```
MCP Server Usage Statistics (last 30 days)
────────────────────────────────────────────────────────────────────────────────

── Global ──
  NAME           CALLS   LAST USED      USAGE
  context7         142   2 hours ago    ████████████████
  puppeteer          0   never          ░░░░░░░░░░░░░░░░  ⚠️ unused

── /Users/xxx/github/my-project ──
  NAME           CALLS   LAST USED      USAGE
  serena            23   1 day ago      ██░░░░░░░░░░░░░░

Total tool calls: 165
```

Options:

- `--period` - Time period for stats (7d, 30d, 90d, all). Default: 30d
- `--sort` - Sort by (calls, name, last-used). Default: calls
- `--json` - Output in JSON format

```bash
# Last 7 days, sorted by name
mcp-tidy stats --period 7d --sort name

# JSON output for scripting
mcp-tidy stats --json
```

### Remove Unused Servers

```bash
mcp-tidy remove
```

Interactive selection (with scope display):

```
? Select servers to remove (enter numbers separated by spaces, or 'all'):
  [1] context7 [global] (142 calls, 2 hours ago)
  [2] puppeteer [global] (0 calls, never used) ⚠️ unused
  [3] serena [/Users/xxx/github/my-project] (23 calls, 1 day ago)

Enter selection: 2

Remove 1 server(s)? [y/N]: y
Backup created: ~/.claude.json.backup.20250105-123456
✓ Removed: puppeteer (from ~/.claude.json)
```

Options:

- `--unused` - Only show unused servers
- `--dry-run` - Preview changes without removing
- `--force` - Remove without confirmation
- `--period` - Period for determining "unused" (7d, 30d, 90d). Default: 30d

```bash
# Preview removal of unused servers
mcp-tidy remove --unused --dry-run

# Force remove all unused servers
mcp-tidy remove --unused --force
```

> **Note**: A timestamped backup (e.g., `~/.claude.json.backup.20250105-123456`) is automatically created before any removal. You can restore it if needed.

## Configuration

mcp-tidy reads from `~/.claude.json` which contains:

- **Global servers**: `mcpServers` key
- **Project servers**: `projects.{path}.mcpServers` key

Usage statistics are collected from Claude Code transcript logs in `~/.claude/projects/`.

## Limitations

- **Claude Code only**: Other MCP clients (Claude Desktop, Cursor, etc.) are not yet supported
- **Config file scope**: Only reads `~/.claude.json`. The following locations are **not** scanned:
  - `.mcp.json` (project-scoped servers added via `claude mcp add --scope project`)
  - `~/.claude/settings.local.json`
  - `.claude/settings.local.json`
  - `managed-mcp.json` (enterprise)
- **Path encoding**: Non-ASCII characters in project paths may not be handled correctly

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request!!

## License

MIT License - see [LICENSE](LICENSE) file.
