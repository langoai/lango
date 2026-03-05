## Context

MCP Plugin System is fully implemented (internal/mcp/, internal/cli/mcp/, TUI settings form) but has zero documentation coverage. All other major features have entries in README.md (Features, CLI Commands, Architecture) and dedicated docs/cli/ pages.

## Goals / Non-Goals

**Goals:**
- Document MCP in README.md Features list, CLI Commands section, and Architecture diagram
- Add MCP to docs/cli/index.md Quick Reference table
- Create docs/cli/mcp.md with full CLI reference matching actual --help output

**Non-Goals:**
- No code changes
- No new MCP features or configuration changes
- No changes to existing automation/p2p/security docs

## Decisions

1. **Follow existing documentation patterns** — Match the style of docs/cli/automation.md (flag tables, examples, tip boxes) for consistency.
2. **Verify against actual CLI** — All documented flags/commands match the real implementation in internal/cli/mcp/*.go.
3. **Place MCP between Workflow and P2P in README** — Follows the logical grouping (automation → integration → networking).
4. **Include Configuration section in mcp.md** — Documents the 3-scope config merge (profile < user < project) and tool naming convention, which are unique to MCP.

## Risks / Trade-offs

- [Docs drift] If MCP CLI flags change, docs need manual update → Mitigated by keeping docs close to --help output patterns
