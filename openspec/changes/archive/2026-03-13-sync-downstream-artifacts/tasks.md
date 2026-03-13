## 1. Prompts Update

- [x] 1.1 Update `prompts/AGENTS.md` — change "fourteen" to "fifteen" tool categories, add Team category
- [x] 1.2 Update `prompts/TOOL_USAGE.md` — add Team Tool section (7 tools), add cron `--timeout` parameter

## 2. README Update

- [x] 2.1 Add 7 new feature bullets to `README.md`
- [x] 2.2 Add `lango status`, `lango onboard --preset`, cron `--timeout` to CLI commands section
- [x] 2.3 Update architecture tree with `internal/p2p/team/`, `internal/p2p/workspace/`, `internal/p2p/gitbundle/`

## 3. Cron Documentation

- [x] 3.1 Add per-job timeout section to `docs/automation/cron.md`
- [x] 3.2 Add idempotent upsert behavior note and `cron.defaultJobTimeout` config

## 4. P2P CLI Documentation

- [x] 4.1 Add incremental git bundle subsection to `docs/cli/p2p.md`
- [x] 4.2 Add task branch management commands to `docs/cli/p2p.md`
- [x] 4.3 Add workflow example for task branch lifecycle

## 5. P2P Network Feature Documentation

- [x] 5.1 Add health monitoring section to `docs/features/p2p-network.md`
- [x] 5.2 Add graceful shutdown section
- [x] 5.3 Add git state divergence detection section
- [x] 5.4 Add reorg protection section
- [x] 5.5 Add event-driven bridges section with 6 bridges

## 6. Economy Feature Documentation

- [x] 6.1 Add Hub V2 section to `docs/features/economy.md`
- [x] 6.2 Add milestone settler section
- [x] 6.3 Add dangling escrow detector section
- [x] 6.4 Update events summary table

## 7. New Documentation Pages

- [x] 7.1 Create `docs/features/config-presets.md` with 4 presets and feature matrices
- [x] 7.2 Create `docs/cli/status.md` with full command reference

## 8. Index and Navigation Updates

- [x] 8.1 Update `docs/getting-started/quickstart.md` with `--preset` flag
- [x] 8.2 Update `docs/cli/index.md` with `lango status` in Quick Reference
- [x] 8.3 Update `docs/features/index.md` with Workspace/Team/Presets cards
- [x] 8.4 Update `docs/index.md` with new feature cards

## 9. Docker and Build

- [x] 9.1 Add `lango-workspaces` volume to `docker-compose.yml`
- [x] 9.2 Add `LANGO_TEAM` and `LANGO_ECONOMY` commented env vars
- [x] 9.3 Add `test-team`, `test-economy`, `test-bridges` targets to `Makefile`
- [x] 9.4 Verify `go build ./...` passes
