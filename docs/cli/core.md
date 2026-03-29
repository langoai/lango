# Core Commands

## lango

Run `lango` without arguments to launch the multi-panel TUI cockpit. This is the default entry point and is equivalent to `lango cockpit`.

```bash
$ lango
```

Only core components (session store, embedding, knowledge) are started. Network and automation components (gateway, channels, cron, MCP, P2P) are not active in TUI mode.

---

## lango cockpit

Launch the multi-panel TUI dashboard explicitly. Same behavior as bare `lango`. The cockpit provides:

- Multi-page layout with Chat, Tools, Status, Sessions, and Settings pages
- Transcript viewport with assistant markdown reflow on resize
- Clear visual separation between user, assistant, status, and approval transcript blocks
- Dedicated turn status strip for ready/streaming/approval/failure states
- Inline tool approval interrupts (`a` allow / `s` allow session / `d` deny)
- Slash commands (`/help`, `/clear`, `/model`, `/status`, `/exit`)
- Key bindings: `Enter` send, `Alt+Enter` newline, `Ctrl+C` cancel/quit, `Ctrl+D` quit

```bash
$ lango cockpit
```

---

## lango chat

Launch the plain chat TUI. A simpler, transcript-first experience without the multi-panel cockpit layout. Suitable for quick interactions that don't require the full dashboard.

```bash
$ lango chat
```

---

## lango serve

Start the gateway server. This boots the full application stack including all enabled channels, tools, embedding, graph, cron, and workflow engines.

```
lango serve
```

The server reads configuration from the active encrypted profile and starts:

- HTTP API on the configured port (default `18789`)
- WebSocket endpoint (if enabled)
- All configured channel adapters (Telegram, Discord, Slack)
- Background systems (cron scheduler, workflow engine) if enabled

Graceful shutdown is handled via `SIGINT` or `SIGTERM` with a 10-second timeout. If shutdown is already in progress, a second `Ctrl+C` forces immediate exit with code `130`.

**Example:**

```bash
$ lango serve
INFO  starting lango  {"version": "0.5.0", "profile": "default"}
INFO  server listening  {"address": ":18789"}
```

---

## lango version

Print the binary version and build timestamp.

```
lango version
```

**Example:**

```bash
$ lango version
lango 0.5.0 (built 2026-02-20T12:00:00Z)
```

---

## lango health

Check whether the gateway server is running and healthy. Sends an HTTP GET to the `/health` endpoint.

```
lango health [--port N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | `18789` | Gateway port to check |

**Examples:**

```bash
# Check default port
$ lango health
ok

# Check custom port
$ lango health --port 9090
ok
```

!!! info
    This command is designed to work as a Docker `HEALTHCHECK`. It exits with code 0 on success and non-zero on failure.

---

## lango onboard

Launch the guided 5-step setup wizard using an interactive TUI. This is the recommended way to configure Lango for the first time.

```
lango onboard [--profile <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Profile name to create or edit |

The wizard walks through five steps:

1. **Provider Setup** -- Choose an AI provider (Anthropic, OpenAI, Gemini, Ollama, GitHub) and enter API credentials
2. **Agent Config** -- Select model (auto-fetched from provider), max tokens, and temperature
3. **Channel Setup** -- Configure Telegram, Discord, or Slack (or skip)
4. **Security & Auth** -- Privacy interceptor, PII redaction, approval policy
5. **Test Config** -- Validate your configuration

All settings are saved to an encrypted profile in `~/.lango/lango.db`.

**Example:**

```bash
# Run with default profile
$ lango onboard

# Create a separate "staging" profile
$ lango onboard --profile staging
```

!!! tip
    For full control over every configuration option, use `lango settings` instead.

---

## lango settings

Open the full interactive configuration editor. Provides access to all configuration options organized by category, including advanced features like embedding, graph, payment, and automation settings.

```
lango settings [--profile <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Profile name to edit |

The settings editor uses a TUI menu interface where you can navigate through categories and edit individual values. Categories are organized into sections:

- **Core:** Providers, Agent, Channels, Tools, Server, Session, Logging, Gatekeeper, Output Manager
- **AI & Knowledge:** Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Librarian, Retrieval, Auto-Adjust, Context Budget, Agent Memory, Multi-Agent, A2A Protocol, Hooks
- **Automation:** Cron Scheduler, Background Tasks, Workflow Engine, RunLedger
- **Payment & Account:** Payment, Smart Account, SA Session Keys, SA Paymaster, SA Modules
- **P2P & Economy:** P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Economy, Risk, Negotiation, Escrow, On-Chain Escrow, Pricing
- **Integrations:** MCP Settings, MCP Server List, Observability
- **Security:** Security, Auth, Security DB Encryption, Security KMS

Press `/` to search across all categories by keyword.

Changes are saved to the active encrypted profile.

!!! note
    This command requires an interactive terminal. For scripted configuration, use `lango config import` with a JSON file.

---

## lango doctor

Run diagnostics to check your Lango configuration and environment for common issues. Optionally attempt to fix problems automatically.

```
lango doctor [--fix] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--fix` | bool | `false` | Attempt to automatically fix issues |
| `--json` | bool | `false` | Output results as JSON |

**Checks performed include:**

- Configuration profile validity
- AI provider configuration and API keys
- API key security (env-var best practices)
- Channel token validation (Telegram, Discord, Slack)
- Session database accessibility
- Server port availability / network configuration
- Security configuration (signer, interceptor, encryption)
- Companion connectivity (WebSocket gateway status)
- Observational memory configuration
- Output scanning and interceptor settings
- Embedding / RAG provider and model setup
- Graph store configuration
- Multi-agent orchestration settings
- Recent multi-agent turn traces (`loop_detected`, `empty_after_tool_use`, `timeout`)
- Persisted isolated-turn leak detection
- A2A protocol connectivity
- RunLedger configuration invariants
- Tool hooks configuration
- Agent registry health
- Librarian status
- Retrieval coordinator and auto-adjust settings
- Context budget allocation ratio validation
- Approval system status
- Economy layer configuration
- Contract configuration
- Observability configuration

**Examples:**

```bash
# Run diagnostics
$ lango doctor

# Attempt auto-fix for known issues
$ lango doctor --fix

# Machine-readable output
$ lango doctor --json
```

When multi-agent runtime failures exist, `--json` now includes structured trace metadata such as `traceFailures[].traceId`, `outcome`, `errorCode`, `causeClass`, and `summary`, plus `isolationLeakCount` when applicable.

!!! tip
    Run `lango doctor` after `lango onboard` to verify your setup is correct. In multi-agent mode, `doctor` also reports recent failed turn traces and whether isolated specialist turns have leaked into persisted parent history.

---

## lango agent trace

Inspect turn traces for diagnostics.

```bash
# List recent failed traces
$ lango agent trace list

# Filter by outcome
$ lango agent trace list --outcome timeout --limit 10

# Filter by session
$ lango agent trace list --session tui-123456

# View detailed event timeline for a trace
$ lango agent trace abc-123-def

# JSON output
$ lango agent trace list --json
```

## lango agent graph

Show the delegation graph for a session, displaying which agents were involved and handoff edges.

```bash
$ lango agent graph <session-key>
$ lango agent graph tui-123456 --json
```

## lango agent trace metrics

Display trace-derived per-agent performance metrics (success rate, turn count, duration percentiles). Distinct from `lango metrics agents` which shows token usage.

```bash
$ lango agent trace metrics
$ lango agent trace metrics --agent operator --json
```

---

## lango config

Configuration profile management. Manage multiple configuration profiles for different environments or setups.

```
lango config <subcommand>
```

### lango config list

List all configuration profiles.

```
lango config list
```

**Output columns:**

| Column | Description |
|--------|-------------|
| NAME | Profile name |
| ACTIVE | `*` if currently active |
| VERSION | Profile version number |
| CREATED | Creation timestamp |
| UPDATED | Last updated timestamp |

**Example:**

```bash
$ lango config list
NAME      ACTIVE  VERSION  CREATED              UPDATED
default   *       3        2026-02-10 08:00:00  2026-02-20 14:30:00
staging           1        2026-02-15 10:00:00  2026-02-15 10:00:00
```

---

### lango config create

Create a new profile with default configuration.

```
lango config create <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name for the new profile |

**Example:**

```bash
$ lango config create staging
Profile "staging" created with default configuration.
```

---

### lango config use

Switch to a different configuration profile.

```
lango config use <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Profile name to activate |

**Example:**

```bash
$ lango config use staging
Switched to profile "staging".
```

---

### lango config delete

Delete a configuration profile. Prompts for confirmation unless `--force` is used.

```
lango config delete <name> [--force]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Profile name to delete |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force`, `-f` | bool | `false` | Skip confirmation prompt |

**Example:**

```bash
$ lango config delete staging
Delete profile "staging"? This cannot be undone. [y/N]: y
Profile "staging" deleted.

$ lango config delete staging --force
Profile "staging" deleted.
```

---

### lango config import

Import and encrypt a JSON configuration file. The source file is deleted after import for security.

```
lango config import <file> [--profile <name>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Path to the JSON configuration file |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Name for the imported profile |

**Example:**

```bash
$ lango config import ./config.json --profile production
Imported "./config.json" as profile "production" (now active).
Source file deleted for security.
```

---

### lango config export

Export a profile as plaintext JSON. Requires passphrase verification via bootstrap.

```
lango config export <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Profile name to export |

**Example:**

```bash
$ lango config export default
WARNING: exported configuration contains sensitive values in plaintext.
{
  "agent": {
    "provider": "anthropic",
    ...
  }
}
```

!!! warning
    The exported JSON contains sensitive values (API keys, tokens) in plaintext. Handle with care.

---

### lango config validate

Validate the active configuration profile.

```
lango config validate
```

**Example:**

```bash
$ lango config validate
Profile "default" configuration is valid.
```
