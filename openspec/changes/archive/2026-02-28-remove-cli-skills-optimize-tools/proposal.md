## Why

The lango agent registers 42 embedded default skills (all `lango <command>` shell wrappers) as tools. These skills spawn `lango` CLI as a subprocess, which requires passphrase authentication and always fails in non-interactive agent mode. The agent attempts these failing skills before using equivalent built-in tools, wasting cycles and confusing error handling.

## What Changes

- **Remove all 42 lango CLI wrapper SKILL.md files** from `skills/` directory (agent-list, config-*, cron-*, graph-*, memory-*, p2p-*, secrets-*, security-*, serve, version, workflow-*, doctor)
- **Preserve embed.go logic** with a `.placeholder/SKILL.md` so future external skills can still be embedded
- **Add tool priority guidance** to `prompts/TOOL_USAGE.md` — new "Tool Selection Priority" section instructing agents to prefer built-in tools over skills
- **Add tool selection note** to `prompts/AGENTS.md` — brief directive reinforcing built-in-first policy
- **Add runtime priority note** in `internal/knowledge/retriever.go` — "Available Skills" section now includes a disclaimer to prefer built-in tools

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `agent-prompting`: Added tool selection priority guidance to TOOL_USAGE.md and AGENTS.md prompts
- `skill-system`: Removed all default embedded CLI wrapper skills; embed.go preserved with placeholder for future use

## Impact

- `skills/` — 42 subdirectories deleted, `.placeholder/SKILL.md` added
- `skills/embed.go` — unchanged (original go:embed logic preserved)
- `prompts/TOOL_USAGE.md` — new "Tool Selection Priority" section prepended
- `prompts/AGENTS.md` — tool selection directive added before knowledge system description
- `internal/knowledge/retriever.go` — skills section in AssemblePrompt now includes priority note
- No API changes, no breaking changes, no dependency changes
