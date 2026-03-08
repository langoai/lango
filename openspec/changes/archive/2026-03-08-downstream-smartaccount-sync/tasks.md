# Tasks: Smart Account Downstream Sync

## WU-1: TUI Smart Account Settings
- [x] 1.1 Create `internal/cli/settings/forms_smartaccount.go` with 4 form constructors
- [x] 1.2 Add Smart Account categories to `internal/cli/settings/menu.go` Infrastructure section
- [x] 1.3 Add case handlers to `internal/cli/settings/editor.go`
- [x] 1.4 Add 19 config mappings to `internal/cli/tuicore/state_update.go`
- [x] 1.5 Verify `go build ./...` and `go test ./...` pass

## WU-2: README.md
- [x] 2.1 Add Smart Accounts feature bullet to features list
- [x] 2.2 Add `lango account` CLI commands to CLI reference section

## WU-3: Feature Documentation
- [x] 3.1 Create `docs/features/smart-accounts.md` with architecture, session keys, paymaster, policy, modules, tools, config

## WU-4: CLI Documentation
- [x] 4.1 Create `docs/cli/smartaccount.md` documenting all 11 CLI commands
- [x] 4.2 Add Smart Account section to `docs/cli/index.md`

## WU-5: Configuration Documentation
- [x] 5.1 Add SmartAccount section to `docs/configuration.md` with all 19 config keys

## WU-6: Tool Usage Documentation
- [x] 6.1 Add Smart Account Tool section to `prompts/TOOL_USAGE.md` with all 12 tools
- [x] 6.2 Add `lango account` to exec tool blocklist

## WU-7: Cross-References
- [x] 7.1 Add Smart Account card and feature status row to `docs/features/index.md`
- [x] 7.2 Add Smart Account Integration section to `docs/features/economy.md`
- [x] 7.3 Add ERC-7579 Module Contracts section to `docs/features/contracts.md`

## WU-8: Build & Deploy
- [x] 8.1 Add `check-abi` target to `Makefile`
- [x] 8.2 Add `LANGO_SMART_ACCOUNT` env var to `docker-compose.yml`

## WU-9: Multi-Agent Routing & Prompts
- [x] 9.1 Add 7 smart account prefixes to vault agent in `internal/orchestration/tools.go`
- [x] 9.2 Add smart account keywords to vault agent
- [x] 9.3 Add 7 entries to `capabilityMap`
- [x] 9.4 Update vault agent Description and Instruction text
- [x] 9.5 Update `prompts/agents/vault/IDENTITY.md` with smart account operations
- [x] 9.6 Add Smart Account tool category to `prompts/AGENTS.md`
- [x] 9.7 Verify `go build ./...` and `go test ./...` pass
