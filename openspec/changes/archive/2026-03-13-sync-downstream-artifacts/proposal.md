## Why

The `dev` branch has accumulated 257 files of core changes (+13,270 lines) adding major features — P2P Workspace, Git Bundle (incremental + task branches), Team Coordination (health monitoring, conflict resolution), Escrow Hub V2, EventMonitor Reorg Protection, Event-Driven Bridges, Cron Enhancements, and CLI Reorganization (presets, status dashboard, onboarding wizard). The downstream artifacts (documentation, prompts, Docker config, Makefile) were not updated to reflect these changes, creating a documentation gap.

## What Changes

- **prompts/AGENTS.md**: Update tool category count from fourteen to fifteen, add Team category
- **prompts/TOOL_USAGE.md**: Add Team Tool section (7 tools), add cron `--timeout` parameter
- **README.md**: Add 7 new feature descriptions, `lango status` command, `--preset` flag, cron `--timeout`, architecture tree updates
- **docs/automation/cron.md**: Add per-job timeout section, upsert behavior, `defaultJobTimeout` config
- **docs/cli/p2p.md**: Add incremental git bundle and task branch management documentation
- **docs/features/p2p-network.md**: Add health monitoring, graceful shutdown, divergence detection, reorg protection, event-driven bridges
- **docs/features/economy.md**: Add Hub V2, milestone settler, dangling detector documentation
- **docs/features/config-presets.md** (NEW): Document 4 config presets with feature matrices
- **docs/cli/status.md** (NEW): Full `lango status` command reference
- **docs/getting-started/quickstart.md**: Add preset selection with `--preset` flag
- **docs/cli/index.md**: Add `lango status` to Quick Reference
- **docs/index.md**, **docs/features/index.md**: Add feature cards for Workspaces, Teams, Presets
- **docker-compose.yml**: Add workspace volume, team/economy env vars
- **Makefile**: Add `test-team`, `test-economy`, `test-bridges` targets

## Capabilities

### New Capabilities

- `downstream-docs-sync`: Synchronization of documentation, prompts, Docker, and Makefile artifacts with core feature changes across 10 work units

### Modified Capabilities

- `p2p-documentation`: Updated with team health, reorg protection, bridges, git bundle docs
- `docs-only`: Updated pattern used for documentation-only changes
- `docker-deployment`: Updated with workspace volumes and team/economy env vars

## Impact

- 13 modified files + 2 new files across docs, prompts, Docker, and Makefile
- No code logic changes — documentation and configuration only
- Go build verified passing
- New Makefile targets for team, economy, and bridge testing
