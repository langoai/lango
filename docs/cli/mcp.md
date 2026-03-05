# MCP Commands

Commands for managing external MCP (Model Context Protocol) server connections. MCP allows Lango to connect to external tool servers, automatically discover their tools, and expose them to the agent.

MCP must be enabled in configuration (`mcp.enabled = true`).

```
lango mcp <subcommand>
```

---

## lango mcp list

List all configured MCP servers with their type, enabled status, and endpoint.

```
lango mcp list
```

**Output columns:**

| Column | Description |
|--------|-------------|
| NAME | Server name |
| TYPE | Transport type (`stdio`, `http`, `sse`) |
| ENABLED | `yes` or `no` |
| ENDPOINT | Command (stdio) or URL (http/sse) |

**Example:**

```bash
$ lango mcp list
NAME            TYPE    ENABLED  ENDPOINT
filesystem      stdio   yes      npx @modelcontextprotocol/server-filesystem
github          http    yes      https://mcp.github.com/v1
slack           sse     yes      https://mcp-slack.example.com/sse
```

---

## lango mcp add

Add a new MCP server configuration.

```
lango mcp add <name> [flags]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name (unique identifier) |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | `stdio` | Transport type: `stdio`, `http`, `sse` |
| `--command` | string | | Executable command (required for `stdio`) |
| `--args` | string | | Comma-separated arguments for the command (`stdio`) |
| `--url` | string | | Endpoint URL (required for `http`/`sse`) |
| `--env` | strings | | Environment variables in `KEY=VALUE` format (repeatable) |
| `--header` | strings | | HTTP headers in `KEY=VALUE` format (repeatable) |
| `--scope` | string | `user` | Config scope: `user` or `project` |
| `--safety` | string | `dangerous` | Safety level: `safe`, `moderate`, `dangerous` |

!!! note "Transport Requirements"
    - `stdio` requires `--command` (the executable to spawn)
    - `http` and `sse` require `--url` (the server endpoint)

**Examples:**

```bash
# Add a stdio-based MCP server
$ lango mcp add filesystem \
    --type stdio \
    --command "npx" \
    --args "@modelcontextprotocol/server-filesystem,/home/user/docs" \
    --scope project
MCP server "filesystem" added (scope: project)

# Add an HTTP-based MCP server with authentication
$ lango mcp add github \
    --type http \
    --url "https://mcp.github.com/v1" \
    --header "Authorization=Bearer ghp_xxxx" \
    --safety moderate
MCP server "github" added (scope: user)

# Add an SSE-based MCP server with environment variables
$ lango mcp add slack \
    --type sse \
    --url "https://mcp-slack.example.com/sse" \
    --env "SLACK_TOKEN=xoxb-xxxx"
MCP server "slack" added (scope: user)
```

---

## lango mcp remove

Remove an MCP server configuration. Aliases: `rm`.

```
lango mcp remove <name> [--scope <scope>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name to remove |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--scope` | string | | Scope to remove from: `user` or `project` (default: search all scopes) |

**Example:**

```bash
$ lango mcp remove filesystem
MCP server "filesystem" removed.

$ lango mcp remove github --scope user
MCP server "github" removed from user scope.
```

---

## lango mcp get

Show detailed information about an MCP server, including its configuration and discovered tools.

```
lango mcp get <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name |

**Example:**

```bash
$ lango mcp get filesystem
Name:       filesystem
Type:       stdio
Enabled:    yes
Safety:     dangerous
Command:    npx
Args:       @modelcontextprotocol/server-filesystem /home/user/docs
Env vars:   0

Tools (3):
  read_file          Read contents of a file
  write_file         Write contents to a file
  list_directory     List files in a directory
```

---

## lango mcp test

Test connectivity to an MCP server. Performs a handshake, measures latency, counts available tools, and pings the session.

```
lango mcp test <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name to test |

**Example:**

```bash
$ lango mcp test filesystem
Testing MCP server "filesystem"...
  Transport:  stdio (npx)
  Handshake:  OK (142ms)
  Tools:      3 available
  Ping:       OK
```

---

## lango mcp enable

Enable a previously disabled MCP server.

```
lango mcp enable <name> [--scope <scope>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name to enable |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--scope` | string | | Scope: `user` or `project` (default: search all scopes) |

**Example:**

```bash
$ lango mcp enable github
MCP server "github" enabled.
```

---

## lango mcp disable

Disable an MCP server without removing its configuration.

```
lango mcp disable <name> [--scope <scope>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Server name to disable |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--scope` | string | | Scope: `user` or `project` (default: search all scopes) |

**Example:**

```bash
$ lango mcp disable slack
MCP server "slack" disabled.
```

---

## Configuration

MCP server configurations are stored in JSON files and merged in priority order:

| Scope | File | Description |
|-------|------|-------------|
| Profile | Active config profile | Base MCP settings (`mcp.enabled`, `mcp.defaultTimeout`) |
| User | `~/.lango/mcp.json` | User-wide server definitions |
| Project | `.lango-mcp.json` | Project-specific server definitions (highest priority) |

When the same server name exists in multiple scopes, the higher-priority scope wins. Use `--scope` flags to target a specific scope when adding, removing, enabling, or disabling servers.

### TUI Settings

MCP servers can also be managed through the interactive TUI settings editor (`lango settings`), which provides forms for adding and configuring servers with transport-specific fields.

### Key Config Options

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mcp.enabled` | bool | `false` | Enable MCP integration |
| `mcp.defaultTimeout` | duration | `30s` | Default server connection timeout |
| `mcp.maxOutputTokens` | int | `25000` | Max output tokens for MCP tool results |
| `mcp.servers.<name>` | object | | Server configuration (set via `lango mcp add` or JSON files) |

### Tool Naming Convention

Tools discovered from MCP servers are registered with the naming pattern:

```
mcp__{serverName}__{toolName}
```

For example, a `read_file` tool from a server named `filesystem` becomes `mcp__filesystem__read_file`.
