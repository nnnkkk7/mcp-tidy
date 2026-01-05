# mcp-tidy

A CLI tool to visualize and manage MCP (Model Context Protocol) server usage in Claude Code.

## Features

- **List** all configured MCP servers (global and project-scoped)
- **Stats** view usage statistics for each server
- **Remove** unused servers interactively with backup support

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

## License

MIT License - see [LICENSE](LICENSE) file.
