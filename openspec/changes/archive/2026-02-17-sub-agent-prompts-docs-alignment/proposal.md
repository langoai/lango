## Why

The automator sub-agent and its 16 automation tools (cron_*, bg_*, workflow_*) were implemented in code but not reflected in prompt files, documentation, or Docker configuration. Sub-agent default prompt files (`prompts/agents/`) were missing entirely for all 7 agents, the global prompts (AGENTS.md, TOOL_USAGE.md) lacked automation tool categories, README listed only 6 sub-agents, and docker-compose.yml had no option for runtime prompt customization.

## What Changes

- Create `prompts/agents/<name>/IDENTITY.md` for all 7 sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler) with content matching `agentSpecs[].Instruction` from `orchestration/tools.go`
- Add Cron, Background, and Workflow tool categories to `prompts/AGENTS.md`
- Add Cron Tool, Background Tool, and Workflow Tool usage sections to `prompts/TOOL_USAGE.md`
- Update README.md to include automator in all 5 sub-agent reference locations (feature list, directory structure, per-agent prompt docs, orchestration table, workflow supported agents)
- Add commented prompts volume mount to `docker-compose.yml` for runtime prompt customization

## Capabilities

### New Capabilities
- `sub-agent-default-prompts`: Default IDENTITY.md prompt files for all 7 sub-agents in `prompts/agents/`

### Modified Capabilities
- `embedded-prompt-files`: Add 3 automation tool categories (Cron, Background, Workflow) to AGENTS.md and TOOL_USAGE.md
- `docker-deployment`: Add optional prompts volume mount to docker-compose.yml
- `multi-agent-orchestration`: Add automator to all README sub-agent reference lists

## Impact

- **Prompt files**: 7 new `prompts/agents/*/IDENTITY.md` files (auto-included in Docker via existing `COPY prompts/`)
- **Global prompts**: `prompts/AGENTS.md` and `prompts/TOOL_USAGE.md` updated with automation documentation
- **README**: 5 locations updated to reflect 7 sub-agents instead of 6
- **Docker**: `docker-compose.yml` gains an optional volume mount comment
- **No Go code changes**: All changes are documentation/prompt files only — build and tests unaffected
