# Agent & Memory

Commands for inspecting agent configuration and managing observational or per-agent memory.

Persistent agent memory stores original content as broker-protected ciphertext. CLI and retrieval flows display decrypted values when available, while SQLite search columns keep only redacted projections.

---

## Agent Commands

### lango agent status

Show the current agent mode and configuration.

```
lango agent status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango agent status
Agent Status
  Mode:              single
  Provider:          anthropic
  Model:             claude-sonnet-4-20250514
  Multi-Agent:       false
  Max Turns:         50
  Error Correction:  true
  A2A Enabled:       false
  P2P Enabled:       false
  Hooks Enabled:     true

Agent Registry
  Builtin Agents:    8
  User Agents:       0
  Active Agents:     8
```

When multi-agent mode is enabled:

```bash
$ lango agent status
Agent Status
  Mode:              multi-agent
  Provider:          anthropic
  Model:             claude-sonnet-4-20250514
  Multi-Agent:       true
  Teammate Runtime:  dynamic-v1
  Max Turns:         75
  Error Correction:  true
  Delegation Rounds: 10
  A2A Enabled:       true
  A2A Base URL:      http://localhost:18789
  A2A Agent:         lango
  Remote Agents:     1
  P2P Enabled:       true
  Hooks Enabled:     true

Agent Registry
  Builtin Agents:    8
  User Agents:       1
  Active Agents:     9
  Agents Dir:        ~/.lango/agents
```

---

### lango agent list

List all available sub-agents (local) and remote A2A agents. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango agent list [--output table|json] [--check]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |
| `--check` | bool | `false` | Test connectivity to remote agents |

**Local agents** are always listed regardless of multi-agent configuration. This includes embedded default agents and any user-defined agents from `agent.agentsDir`. The `SOURCE` column shows where each definition came from (`embedded`, `user`, or `builtin` when present).

Missing `agent.agentsDir` paths are treated as no user-defined agents. If the configured directory contains an `AGENT.md` file that cannot be read or parsed, `lango agent list` and `lango agent status` fail with an error that includes the affected file path instead of silently omitting user agents or reporting partial registry counts.

Representative built-in local agents:

| Agent | Source | Description |
|-------|--------|-------------|
| automator | embedded | Automation: cron scheduling, background tasks, workflow orchestration |
| librarian | embedded | Knowledge management: search, RAG, graph traversal, skills, inquiries |
| navigator | embedded | Web browsing: page navigation, interaction, screenshots |
| vault | embedded | Security operations: encryption, secret management, payments |

**Example:**

```bash
$ lango agent list
NAME       SOURCE    DESCRIPTION
automator  embedded  Automation: cron scheduling, background tasks, workflow...
librarian  embedded  Knowledge management: search, RAG, graph traversal...
navigator  embedded  Web browsing: page navigation, interaction, screenshots
vault      embedded  Security operations: encryption, secret management...

NAME           SOURCE  URL                                             STATUS
weather-agent  a2a     http://weather-svc:8080/.well-known/agent.json  ok
```

Use `--check` to verify remote agent connectivity:

```bash
$ lango agent list --check
# Remote rows show "ok", "unreachable", or an HTTP status code in the STATUS column.
```

---

### lango agent tools

Show tool category availability derived from the active configuration. This command does not list individual tools or tool-to-agent assignments. It writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango agent tools [--output table|json] [--category <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |
| `--category` | string | `""` | Filter to a single tool category |

**Example:**

```bash
$ lango agent tools
CATEGORY      ENABLED  DESCRIPTION
exec          yes      Shell command execution
filesystem    yes      File system operations
browser       no       Web browsing
crypto        yes      Cryptographic operations
payment       yes      Blockchain payments (USDC on Base)
```

Enabled/disabled values come from the current config profile. Use `--output json` to inspect `name`, `description`, optional `config_key`, and `enabled` fields programmatically.

Operator execution tools validate their wrapper inputs directly: `exec` and `exec_bg` require `command`, while `exec_status` and `exec_stop` require the background-process `id`.

Navigator browser interactions use action-specific validation: `browser_action` requires `selector` for `click`, `get_text`, `get_element_info`, and `wait`; `type` requires both `selector` and `text`; `eval` requires JavaScript in `text`.
Navigator search/navigation entrypoints also validate top-level inputs directly: `browser_search` requires `query`, and `browser_navigate` requires `url`.

Vault security tools validate required inputs at the tool boundary: `crypto_encrypt`, `crypto_sign`, and `crypto_hash` require `data`; `crypto_decrypt` requires `ciphertext`; `secrets_store` requires `name` and `value`; `secrets_get` and `secrets_delete` require `name`.

---

### lango agent hooks

Show registered tool hooks (middleware) in the tool execution chain.
The command writes through the Cobra command output stream so wrappers and test harnesses can capture text or JSON output directly.

```
lango agent hooks [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango agent hooks
Hook Configuration
  Enabled:          true
  Security Filter:  true
  Access Control:   true
  Event Publishing: true
  Knowledge Save:   true
  Blocked Commands: rm -rf /

Registered Hooks
  Pre-hooks:
    security_filter           priority=100  wirable=true
    agent_access_control      priority=200  wirable=true
    eventbus                  wirable=false  (requires a running event bus (unavailable in CLI mode))
  Post-hooks:
    knowledge_save            priority=100  wirable=true
      saveableTools: search_knowledge, search_learnings, graph_traverse
```

Use `--output json` to inspect the same information programmatically, including `registry.preHooks`, `registry.postHooks`, and hook-specific `details`.

---

## Memory Commands

Manage [observational memory](../features/observational-memory.md) entries. Memory commands require a `--session` flag to scope operations to a specific session.

### lango memory list

List observations and reflections for a session. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango memory list --session <key> [--type <type>] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--session` | string | *required* | Session key to query |
| `--type` | string | (all) | Filter by type: `observations` or `reflections` |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango memory list --session user-123
ID        TYPE          TOKENS  CREATED           CONTENT
a1b2c3d4  observation   45      2026-02-20 14:30  User prefers concise answers and dislikes...
e5f6g7h8  reflection    120     2026-02-20 14:35  The user has shown a consistent pattern of...
```

---

### lango memory status

Show observational memory status and configuration for a session. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango memory status --session <key> [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--session` | string | *required* | Session key to query |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango memory status --session user-123
Observational Memory Status (session: user-123)
  Enabled:                      true
  Provider:                     anthropic
  Model:                        claude-haiku-4-5-20251001
  Observations:                 12 (540 tokens)
  Reflections:                  3 (360 tokens)
  Message Token Threshold:      1000
  Observation Token Threshold:  2000
  Max Message Token Budget:     8000
```

---

### lango memory clear

Clear all observations and reflections for a session. Prompts for confirmation unless `--force` is specified. The command uses Cobra command streams for both prompt output and confirmation input, so wrappers and test harnesses can drive the interaction through `cmd.OutOrStdout()` and `cmd.InOrStdin()`.

```
lango memory clear <session-key> [--force]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `session-key` | Yes | Session key to clear |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt |

**Example:**

```bash
$ lango memory clear user-123
This will delete all observations and reflections for session 'user-123'.
Continue? [y/N]: y
Cleared all memory entries for session 'user-123'.
```

!!! warning
    This operation is irreversible. All observations and reflections for the session will be permanently deleted.

---

## Graph Commands

Graph-store commands now live in the dedicated [Graph CLI Reference](graph.md).
Use that page for `lango graph status/query/stats/clear/add/export/import`.

---

## Agent Memory Commands

Manage per-agent persistent memory. Agent memory enables cross-session context retention for individual sub-agents. Requires `agentMemory.enabled: true`.

### lango memory agents

List all agents that have persistent memory entries. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango memory agents [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango memory agents
AGENT           ENTRIES  LAST UPDATED
librarian       45       2026-03-01 14:30
operator        23       2026-03-01 12:15
navigator       12       2026-02-28 09:00
```

---

### lango memory agent

Show memory entries for a specific agent. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango memory agent <name> [--limit N] [--output table|json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Agent name |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `20` | Maximum entries to show |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango memory agent librarian
ID        CONTENT                                              USES  CREATED
a1b2c3d4  User prefers concise code examples in Go             8     2026-02-20 14:30
e5f6g7h8  Project uses BoltDB for graph storage                5     2026-02-22 10:15
i9j0k1l2  User works with microservices architecture           3     2026-02-25 16:45
```
