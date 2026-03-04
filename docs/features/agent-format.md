---
title: AGENT.md File Format
---

# AGENT.md File Format

Custom agents are defined using `AGENT.md` files — markdown documents with YAML frontmatter. This format enables declarative agent definition with rich instruction bodies.

## File Structure

An `AGENT.md` file consists of two sections:

1. **YAML frontmatter** — Agent metadata enclosed between `---` delimiters
2. **Markdown body** — The agent's system instruction

```markdown
---
name: code-reviewer
description: Reviews code for quality, security, and best practices
status: active
prefixes:
  - review_*
  - lint_*
keywords:
  - review
  - code quality
  - security audit
capabilities:
  - code-review
  - security-analysis
accepts: "code snippets, file paths, repository URLs"
returns: "structured review with severity ratings"
cannot_do:
  - execute code
  - modify files
always_include: false
session_isolation: false
---

You are a code review specialist. Analyze code for quality, security vulnerabilities,
and adherence to best practices. Provide structured feedback with severity ratings.

## Review Format

For each issue found:
1. **Location** — File and line reference
2. **Severity** — Critical, Major, Minor, or Suggestion
3. **Description** — Clear explanation of the issue
4. **Fix** — Concrete code suggestion
```

## Frontmatter Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique agent name (used for routing and identification) |
| `description` | string | No | `""` | Short description shown in agent listings |
| `status` | string | No | `active` | Agent status: `active`, `disabled`, or `draft` |
| `prefixes` | []string | No | `[]` | Tool name prefixes this agent handles (e.g., `review_*`) |
| `keywords` | []string | No | `[]` | Keywords for orchestrator routing decisions |
| `capabilities` | []string | No | `[]` | Capability tags for discovery and matching |
| `accepts` | string | No | `""` | Description of what input this agent accepts |
| `returns` | string | No | `""` | Description of what this agent returns |
| `cannot_do` | []string | No | `[]` | Explicit list of things this agent cannot do |
| `always_include` | bool | No | `false` | Always include in the agent tree even with no matching tools |
| `session_isolation` | bool | No | `false` | Run in isolated child sessions |

### Status Values

| Status | Behavior |
|--------|----------|
| `active` | Agent is loaded and participates in routing |
| `disabled` | Agent is registered but skipped during routing |
| `draft` | Agent is not loaded (work-in-progress) |

## Directory Structure

Agent definitions are loaded from the directory specified by `agent.agentsDir` in configuration. Each agent resides in its own subdirectory:

```
~/.lango/agents/
├── code-reviewer/
│   └── AGENT.md
├── translator/
│   └── AGENT.md
└── data-analyst/
    └── AGENT.md
```

## Loading Sources

Agent definitions are loaded from multiple sources with the following priority:

| Priority | Source | Description |
|----------|--------|-------------|
| 1 | Built-in | Hardcoded agents (operator, navigator, vault, librarian, etc.) |
| 2 | Embedded | Default agents bundled in the binary (`defaults/` directory) |
| 3 | User | User-defined agents from `agent.agentsDir` |
| 4 | Remote | Agents loaded from P2P network |

Higher-priority sources take precedence. User-defined agents cannot override built-in agent names.

## Rendering

Agent definitions can be rendered back to the `AGENT.md` format using `RenderAgentMD()`. The rendered output preserves the frontmatter/body structure:

```
---
name: my-agent
description: My custom agent
status: active
---

Agent instruction body here.
```

## Integration with Multi-Agent Orchestration

When multi-agent mode is enabled (`agent.multiAgent: true`), custom agents are integrated into the orchestrator's routing table alongside built-in agents. The orchestrator uses:

- **prefixes** — to partition tools among agents
- **keywords** — to route user requests by topic affinity
- **capabilities** — to match agents against task requirements
- **cannot_do** — to verify agent suitability and handle rejections

See [Multi-Agent Orchestration](multi-agent.md) for routing details.

## CLI Commands

```bash
lango agent list              # List all agents (built-in + user-defined)
lango agent tools             # Show tool-to-agent assignments
```

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `agent.agentsDir` | `""` | Directory containing user-defined AGENT.md files |
| `agent.multiAgent` | `false` | Enable multi-agent orchestration |
