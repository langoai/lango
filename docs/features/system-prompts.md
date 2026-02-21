# System Prompts

Lango ships with production-quality default prompts embedded directly in the binary. No configuration is needed to get started -- the agent works out of the box with sensible defaults for identity, safety, conversation rules, and tool usage.

## Built-in Prompt Sections

The system prompt is assembled from four prioritized sections. Lower priority numbers appear first in the final prompt.

| File | Section | Priority | Description |
|------|---------|----------|-------------|
| `AGENTS.md` | Identity | 100 | Agent name, role, tool capabilities, knowledge system |
| `SAFETY.md` | Safety | 200 | Secret protection, destructive operation confirmation, PII handling |
| `CONVERSATION_RULES.md` | Conversation Rules | 300 | Anti-repetition rules, channel limits, consistency |
| `TOOL_USAGE.md` | Tool Usage | 400 | Per-tool guidelines for exec, filesystem, browser, crypto, secrets, skills |

The prompt builder sorts all sections by priority and joins their rendered content into a single system prompt string.

## Customizing Prompts

To override the built-in defaults, create a directory with `.md` files matching the section names and set `agent.promptsDir` in your configuration.

```bash
mkdir -p ~/.lango/prompts
echo "You are a helpful coding assistant." > ~/.lango/prompts/AGENTS.md
```

```yaml
# config.yaml
agent:
  promptsDir: "~/.lango/prompts"
```

### Precedence

The system resolves prompts in the following order (first match wins):

1. **`promptsDir`** (directory) -- files from the configured prompts directory override matching built-in sections
2. **Built-in defaults** -- embedded prompts compiled into the binary

When a file in the prompts directory matches a known filename (e.g., `SAFETY.md`), it replaces the corresponding built-in section entirely (last-writer-wins). Unknown `.md` files are added as **custom sections** with priority 900+.

## Adding Custom Sections

Any `.md` file placed in the prompts directory that does not match a known section name is automatically added as a custom section:

```bash
echo "Always respond in formal English." > ~/.lango/prompts/STYLE_GUIDE.md
```

This creates a section with:

- **ID**: `custom_style_guide`
- **Title**: `STYLE GUIDE` (derived from filename, underscores become spaces)
- **Priority**: 900+ (appended after all built-in sections)

Multiple custom sections are ordered by filesystem read order within the 900+ range.

## Per-Agent Customization

In [multi-agent mode](../architecture/index.md), each sub-agent can have its own prompt overrides. Create an `agents/<name>/` subdirectory inside your prompts directory.

### Supported Per-Agent Files

| File | Section | Priority | Behavior |
|------|---------|----------|----------|
| `IDENTITY.md` | Agent Identity | 150 | Replaces the default role description for this agent |
| `SAFETY.md` | Safety | 200 | Overrides shared safety guidelines for this agent |
| `CONVERSATION_RULES.md` | Conversation Rules | 300 | Overrides shared conversation rules for this agent |
| `*.md` (other) | Custom | 900+ | Additional custom sections for this agent |

The per-agent identity section (priority 150) sits between the global identity (100) and safety (200), allowing fine-grained layering.

### Directory Layout

```
~/.lango/prompts/
  AGENTS.md                    # Global identity (priority 100)
  SAFETY.md                    # Global safety (priority 200)
  CONVERSATION_RULES.md        # Global conversation rules (priority 300)
  TOOL_USAGE.md                # Global tool usage (priority 400)
  STYLE_GUIDE.md               # Custom section (priority 900)
  agents/
    operator/
      IDENTITY.md              # Operator-specific identity (priority 150)
      SAFETY.md                # Operator-specific safety override
    navigator/
      IDENTITY.md              # Navigator-specific identity
    planner/
      IDENTITY.md              # Planner-specific identity
      PLANNING_RULES.md        # Custom section for planner only
```

### How Per-Agent Loading Works

1. The shared prompt builder is created from the global prompts directory (or built-in defaults)
2. For each sub-agent, the builder is **cloned**
3. Per-agent files from `agents/<name>/` are loaded and override matching sections in the clone
4. Each agent ends up with its own tailored system prompt

This means per-agent overrides only affect that specific agent -- other agents continue using the shared base prompts.

## Architecture

The prompt system is implemented in `internal/prompt/`:

- **`section.go`** -- defines `SectionID` constants and the `PromptSection` interface
- **`builder.go`** -- `Builder` assembles sections, supports `Add`, `Remove`, `Clone`, and `Build`
- **`defaults.go`** -- `DefaultBuilder()` loads the four built-in sections from embedded files
- **`loader.go`** -- `LoadFromDir()` and `LoadAgentFromDir()` handle file-based overrides
