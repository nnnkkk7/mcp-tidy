# mcp-tidy

> Keep your Claude Code MCP configuration clean — like `go mod tidy` for MCP servers.

[![CI](https://github.com/nnnkkk7/mcp-tidy/actions/workflows/ci.yaml/badge.svg)](https://github.com/nnnkkk7/mcp-tidy/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/mcp-tidy)](https://goreportcard.com/report/github.com/nnnkkk7/mcp-tidy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<!-- TODO: Add terminal GIF demo here using VHS or similar tool -->

## Why mcp-tidy?

### The Problem: MCP Server Bloat

As you use Claude Code, MCP servers accumulate in your `~/.claude.json`. This causes real problems:

| Issue | Impact |
|-------|--------|
| **Context window consumption** | Each server's tool definitions consume 5,000–15,000 tokens at session start. 7 servers can consume 67k tokens (33% of context). |
| **Degraded tool selection** | More tools = higher chance Claude picks the wrong one, especially with similar names. |
| **Slower startup** | Each server needs initialization, adding latency proportional to server count. |

### The Solution

**mcp-tidy** helps you reclaim context and improve Claude Code performance:

- **See** what MCP servers are configured (global and per-project)
- **Understand** which ones you actually use (with call statistics)
- **Clean up** unused servers safely (with automatic backups)

> Removing unused MCP servers can recover 10-30% of your context window.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
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

Output:

```
MCP Server Usage Statistics (last 30 days)

  SERVER     CALLS  LAST USED      USAGE
  context7   142    2 hours ago    ████████████████
  serena     23     1 day ago      ██░░░░░░░░░░░░░░
  puppeteer  0      never          unused

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

Interactive selection:

```
Select servers to remove:

  [1] puppeteer  (unused, 0 calls)
  [2] old-server (last used 45 days ago)

Enter numbers (comma-separated) or 'all': 1

Remove 1 server(s)? [y/N]: y
Backup created: ~/.claude.json.backup.20250105-123456
Removed: puppeteer
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

## Configuration

mcp-tidy reads from `~/.claude.json` which contains:

- **Global servers**: `mcpServers` key
- **Project servers**: `projects.{path}.mcpServers` key

Usage statistics are collected from Claude Code transcript logs in `~/.claude/projects/`.

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-cover

# Build
make build

# Install
make install
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Before submitting, please run:

```bash
make ci  # Run lint + test
```

## License

MIT License - see [LICENSE](LICENSE) file.
