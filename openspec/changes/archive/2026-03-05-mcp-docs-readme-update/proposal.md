## Why

MCP Plugin System (Phase 1-4) implementation and TUI Settings form are complete, but README.md and docs have no MCP-related documentation. All other features (Cron, P2P, Security, etc.) are documented in README Features list, CLI Commands section, Architecture diagram, and `docs/cli/`. MCP is the only undocumented feature.

## What Changes

- Add MCP to README.md Features list
- Add `lango mcp` CLI commands block to README.md CLI Commands section
- Add `mcp/` entries to README.md Architecture diagram (both cli and internal)
- Add MCP Servers section to `docs/cli/index.md` Quick Reference table
- Create new `docs/cli/mcp.md` with full MCP CLI reference documentation

## Capabilities

### New Capabilities

(none — this is a documentation-only change, no new code capabilities)

### Modified Capabilities

- `mcp-integration`: Adding documentation coverage for the existing MCP CLI commands and configuration

## Impact

- `README.md` — Features, CLI Commands, Architecture sections updated
- `docs/cli/index.md` — Quick Reference table updated
- `docs/cli/mcp.md` — New file with full CLI reference
- No code changes, no API changes, no dependency changes
