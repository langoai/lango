# Agent & Memory

Commands for inspecting agent configuration, managing observational memory, and interacting with the knowledge graph store.

Persistent agent memory stores original content as broker-protected ciphertext. CLI and retrieval flows display decrypted values when available, while SQLite search columns keep only redacted projections.

---

## Agent Commands

### lango agent status

Show the current agent mode and configuration.

```
lango agent status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango agent status
Agent Status
  Mode:         single
  Provider:     anthropic
  Model:        claude-sonnet-4-20250514
  Multi-Agent:  false
  A2A Enabled:  false
```

When multi-agent mode is enabled:

```bash
$ lango agent status
Agent Status
  Mode:         multi-agent
  Provider:     anthropic
  Model:        claude-sonnet-4-20250514
  Multi-Agent:  true
  A2A Enabled:  true
  A2A Base URL: http://localhost:18789
  A2A Agent:    lango
```

---

### lango agent list

List all available sub-agents (local) and remote A2A agents.

```
lango agent list [--json] [--check]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |
| `--check` | bool | `false` | Test connectivity to remote agents |

**Local agents** are always listed regardless of multi-agent configuration. This includes built-in agents, embedded default agents, and user-defined agents from `agent.agentsDir`:

| Agent | Description |
|-------|-------------|
| executor | Executes tools including shell commands, file operations, browser automation |
| researcher | Searches knowledge bases, performs RAG retrieval, graph traversal |
| planner | Decomposes complex tasks into steps and designs execution plans |
| memory-manager | Manages conversational memory including observations, reflections |

**Example:**

```bash
$ lango agent list
NAME              TYPE   DESCRIPTION
executor          local  Executes tools including shell commands, file operations, browser automation
researcher        local  Searches knowledge bases, performs RAG retrieval, graph traversal
planner           local  Decomposes complex tasks into steps and designs execution plans
memory-manager    local  Manages conversational memory including observations, reflections

NAME              TYPE    URL                                    STATUS
weather-agent     remote  http://weather-svc:8080/.well-known/agent.json  ok
```

Use `--check` to verify remote agent connectivity:

```bash
$ lango agent list --check
# Remote agents will show "ok", "unreachable", or HTTP status codes
```

---

### lango agent tools

Show tool-to-agent assignments. Displays how tools are partitioned across sub-agents in multi-agent mode.

```
lango agent tools [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango agent tools
AGENT           TOOLS
operator        exec_shell, exec_command, fs_read, fs_write, fs_delete, skill_run
navigator       browser_navigate, browser_search, browser_observe, browser_extract, browser_action, browser_screenshot
vault           crypto_encrypt, crypto_decrypt, secrets_set, payment_send
librarian       search_knowledge, rag_query, graph_traverse, save_knowledge
automator       cron_add, cron_list, bg_submit, workflow_run
chronicler      memory_observe, memory_reflect, memory_recall
(unmatched)     custom_tool_1, custom_tool_2
```

---

### lango agent hooks

Show registered tool hooks (middleware) in the tool execution chain.

```
lango agent hooks [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango agent hooks
HOOK                    TYPE          STATUS
security_filter         pre-execute   active
approval_gate           pre-execute   active
learning_observer       post-execute  active
knowledge_saver         post-execute  active
event_publisher         post-execute  active
```

---

## Memory Commands

Manage [observational memory](../features/observational-memory.md) entries. Memory commands require a `--session` flag to scope operations to a specific session.

### lango memory list

List observations and reflections for a session.

```
lango memory list --session <key> [--type <type>] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--session` | string | *required* | Session key to query |
| `--type` | string | (all) | Filter by type: `observations` or `reflections` |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango memory list --session user-123
ID        TYPE          TOKENS  CREATED           CONTENT
a1b2c3d4  observation   45      2026-02-20 14:30  User prefers concise answers and dislikes...
e5f6g7h8  reflection    120     2026-02-20 14:35  The user has shown a consistent pattern of...
```

---

### lango memory status

Show observational memory status and configuration for a session.

```
lango memory status --session <key> [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--session` | string | *required* | Session key to query |
| `--json` | bool | `false` | Output as JSON |

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

Clear all observations and reflections for a session. Prompts for confirmation unless `--force` is specified.

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
Continue? [y/N] y
Cleared all memory entries for session 'user-123'.
```

!!! warning
    This operation is irreversible. All observations and reflections for the session will be permanently deleted.

---

## Graph Commands

Manage the [knowledge graph](../features/knowledge-graph.md) store. The graph must be enabled in configuration (`graph.enabled = true`).

### lango graph status

Show knowledge graph status and basic information.

```
lango graph status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango graph status
Knowledge Graph Status
  Enabled:       true
  Backend:       bolt
  Database Path: /home/user/.lango/graph.db
  Triples:       1523
```

When the graph is disabled:

```bash
$ lango graph status
Knowledge Graph Status
  Enabled:  false
```

---

### lango graph query

Query triples from the knowledge graph by subject, predicate, and/or object.

```
lango graph query [--subject <s>] [--predicate <p>] [--object <o>] [--limit N] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--subject` | string | | Filter by subject |
| `--predicate` | string | | Filter by predicate (requires `--subject`) |
| `--object` | string | | Filter by object |
| `--limit` | int | `0` | Limit number of results (0 = unlimited) |
| `--json` | bool | `false` | Output as JSON |

!!! note "Query Requirements"
    At least one of `--subject` or `--object` is required. The `--predicate` flag can only be used together with `--subject`.

**Examples:**

```bash
# Query by subject
$ lango graph query --subject "Go"
SUBJECT  PREDICATE    OBJECT
Go       is_a         programming_language
Go       created_by   Google
Go       has_feature  goroutines

# Query by subject and predicate
$ lango graph query --subject "Go" --predicate "has_feature"
SUBJECT  PREDICATE    OBJECT
Go       has_feature  goroutines
Go       has_feature  channels
Go       has_feature  garbage_collection

# Query by object
$ lango graph query --object "Google" --limit 5

# JSON output
$ lango graph query --subject "Go" --json
```

---

### lango graph stats

Show knowledge graph statistics including total triple count and predicate distribution.

```
lango graph stats [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango graph stats
Knowledge Graph Statistics
  Total Triples: 1523

PREDICATE       COUNT
is_a            423
has_feature     312
related_to      289
created_by      156
```

---

### lango graph clear

Clear all triples from the knowledge graph. Prompts for confirmation unless `--force` is specified.

```
lango graph clear [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt |

**Example:**

```bash
$ lango graph clear
This will delete all triples from the knowledge graph.
Continue? [y/N] y
Cleared all triples from the knowledge graph.
```

!!! danger
    This operation is irreversible. All graph data will be permanently deleted.

---

### lango graph add

Add a triple to the knowledge graph.

```
lango graph add --subject <s> --predicate <p> --object <o>
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--subject` | string | *required* | Triple subject |
| `--predicate` | string | *required* | Triple predicate (relationship) |
| `--object` | string | *required* | Triple object |

**Example:**

```bash
$ lango graph add --subject "Go" --predicate "is_a" --object "programming_language"
Triple added: Go → is_a → programming_language
```

---

### lango graph export

Export graph data to a file.

```
lango graph export <file> [--format <format>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Output file path |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `json` | Export format: `json` or `ntriples` |

**Example:**

```bash
$ lango graph export ./graph-backup.json
Exported 1523 triples to ./graph-backup.json
```

---

### lango graph import

Import graph data from a file.

```
lango graph import <file> [--format <format>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Input file path |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `json` | Import format: `json` or `ntriples` |

**Example:**

```bash
$ lango graph import ./graph-backup.json
Imported 1523 triples from ./graph-backup.json
```

---

## Agent Memory Commands

Manage per-agent persistent memory. Agent memory enables cross-session context retention for individual sub-agents. Requires `agentMemory.enabled: true`.

### lango memory agents

List all agents that have persistent memory entries.

```
lango memory agents [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

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

Show memory entries for a specific agent.

```
lango memory agent <name> [--limit N] [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Agent name |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `20` | Maximum entries to show |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango memory agent librarian
ID        CONTENT                                              USES  CREATED
a1b2c3d4  User prefers concise code examples in Go             8     2026-02-20 14:30
e5f6g7h8  Project uses BoltDB for graph storage                5     2026-02-22 10:15
i9j0k1l2  User works with microservices architecture           3     2026-02-25 16:45
```
