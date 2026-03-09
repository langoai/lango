## Why

The `feature/p2p-economy` branch added 3 major feature areas (economy layer, contract interaction, observability) with 164 files and +18,166 lines of backend code, CLI commands, agent tools, and config types. However, zero downstream artifacts were updated — no docs, no prompts, no README sections, no TUI settings forms, and no doctor checks exist for these features. Users cannot discover, configure, or validate these features without documentation and UI support.

## What Changes

- Create 3 new feature documentation pages (economy, contracts, observability)
- Create 3 new CLI reference pages (economy, contract, metrics commands)
- Update feature index, CLI index, configuration reference, and mkdocs navigation
- Add Economy Tool (13 tools) and Contract Tool (3 tools) sections to TOOL_USAGE.md prompt
- Update AGENTS.md tool category count and add 3 new categories
- Update README.md with features, CLI commands, and architecture tree entries
- Create TUI settings forms for Economy (5 sub-forms) and Observability (1 form)
- Wire forms into settings menu and editor with state update handlers
- Create 3 doctor health checks (Economy, Contract, Observability) and register them

## Capabilities

### New Capabilities

_None — this change syncs existing documentation and UI artifacts with already-implemented backend capabilities._

### Modified Capabilities

- `economy-cli`: Add documentation for economy CLI commands
- `contract-interaction`: Add feature and CLI documentation
- `observability`: Add feature and CLI documentation, TUI settings form
- `cli-settings`: Add Economy section (5 forms) and Observability form to TUI settings editor
- `cli-doctor`: Add 3 new health checks (Economy, Contract, Observability)
- `p2p-agent-prompts`: Update TOOL_USAGE.md and AGENTS.md with economy/contract/observability tools
- `mkdocs-documentation-site`: Add 6 new pages to navigation
- `cli-reference`: Add Economy, Contract, and Metrics command groups to CLI index

## Impact

- **Docs**: 6 new markdown files, 4 edited files in `docs/`
- **Prompts**: 2 edited files in `prompts/`
- **README**: 1 edited file (features, CLI, architecture)
- **TUI**: 2 new Go files (`forms_economy.go`, `forms_observability.go`), 3 edited Go files (`menu.go`, `editor.go`, `state_update.go`)
- **Doctor**: 3 new Go files (`economy.go`, `contract.go`, `observability.go`), 1 edited Go file (`checks.go`)
- **Nav**: `mkdocs.yml` updated with 6 new navigation entries
- **Config docs**: `configuration.md` updated with Economy and Observability sections
