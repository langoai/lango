## 1. P2P Teams & ZKP CLI

- [x] 1.1 Create `internal/cli/p2p/team.go` with team list/status/disband subcommands (bootLoader, config check only)
- [x] 1.2 Create `internal/cli/p2p/zkp.go` with zkp status (bootLoader) and circuits (no loader) subcommands
- [x] 1.3 Register team and zkp subcommands in `internal/cli/p2p/p2p.go`
- [x] 1.4 Update `internal/cli/p2p/p2p_test.go` subcommand count (11→13)

## 2. Agent CLI Enhancements

- [x] 2.1 Create `internal/cli/agent/catalog.go` with `agent tools [--category]` command (cfgLoader)
- [x] 2.2 Create `internal/cli/agent/hooks.go` with `agent hooks` command (cfgLoader)
- [x] 2.3 Register tools and hooks subcommands in `internal/cli/agent/agent.go`

## 3. New Top-Level CLI Commands

- [x] 3.1 Create `internal/cli/a2a/` package with a2a.go, card.go, check.go (a2a card + a2a check with 1MB LimitReader)
- [x] 3.2 Create `internal/cli/learning/` package with learning.go, status.go (cfgLoader), history.go (bootLoader, uses toolchain.Truncate)
- [x] 3.3 Create `internal/cli/librarian/` package with librarian.go, status.go (cfgLoader), inquiries.go (bootLoader, uses toolchain.Truncate)
- [x] 3.4 Create `internal/cli/approval/` package with approval.go, status.go (bootLoader)
- [x] 3.5 Register all 4 new commands in `cmd/lango/main.go` with proper group IDs and both cfgLoader/bootLoader where needed

## 4. Graph Store Extended CLI

- [x] 4.1 Add `AllTriples()` to `graph.Store` interface and implement in `bolt_store.go` (SPO bucket scan)
- [x] 4.2 Create `internal/cli/graph/add.go` with MarkFlagRequired for subject/predicate/object
- [x] 4.3 Create `internal/cli/graph/export.go` with JSON/CSV format support
- [x] 4.4 Create `internal/cli/graph/import_cmd.go` for JSON triple import
- [x] 4.5 Register add/export/import subcommands in `internal/cli/graph/graph.go`

## 5. Memory, Payment, Workflow Additions

- [x] 5.1 Add `ListAgentNames()`/`ListAll()` to `agentmemory.Store` interface and implement in `mem_store.go`
- [x] 5.2 Create `internal/cli/memory/agent_memory.go` with `memory agents` and `memory agent <name>` commands
- [x] 5.3 Create `internal/cli/payment/x402.go` with x402 config display (no redundant Status field)
- [x] 5.4 Create `internal/cli/workflow/validate.go` with YAML workflow validation
- [x] 5.5 Register new subcommands in respective parent command files
- [x] 5.6 Update `internal/cli/payment/payment_test.go` subcommand count (5→6)

## 6. TUI Settings Enhancements

- [x] 6.1 Create `internal/cli/settings/forms_hooks.go` with NewHooksForm() and NewAgentMemoryForm()
- [x] 6.2 Modify `forms_agent.go` to add maxDelegationRounds, maxTurns, errorCorrectionEnabled, agentsDir fields
- [x] 6.3 Modify `forms_knowledge.go` to add librarian fields and skill import fields
- [x] 6.4 Add "hooks" and "agent_memory" cases in `editor.go` handleMenuSelection
- [x] 6.5 Add "Hooks" and "Agent Memory" menu items in `menu.go`
- [x] 6.6 Add 11 new field key cases in `tuicore/state_update.go`

## 7. Doctor & Onboard Enhancements

- [x] 7.1 Create `internal/cli/doctor/checks/tool_hooks.go` (ToolHooksCheck)
- [x] 7.2 Create `internal/cli/doctor/checks/agent_registry.go` (AgentRegistryCheck)
- [x] 7.3 Create `internal/cli/doctor/checks/librarian.go` (LibrarianCheck)
- [x] 7.4 Create `internal/cli/doctor/checks/approval.go` (ApprovalCheck)
- [x] 7.5 Register 4 new checks in `internal/cli/doctor/checks/checks.go` AllChecks()
- [x] 7.6 Add advanced feature hints to `internal/cli/onboard/onboard.go`

## 8. Documentation

- [x] 8.1 Create `docs/features/agent-format.md` (AGENT.md file format specification)
- [x] 8.2 Create `docs/features/learning.md` (learning system overview)
- [x] 8.3 Create `docs/features/zkp.md` (ZKP system overview)
- [x] 8.4 Create `docs/security/approval-cli.md` (approval system docs)
- [x] 8.5 Create `docs/cli/a2a.md` and `docs/cli/learning.md` (CLI reference)
- [x] 8.6 Update `docs/cli/index.md` with all new commands
- [x] 8.7 Update `docs/cli/p2p.md`, `docs/cli/agent-memory.md`, `docs/cli/payment.md`, `docs/cli/automation.md`
- [x] 8.8 Update `docs/features/multi-agent.md` and `docs/features/p2p-network.md`
- [x] 8.9 Update `docs/cli/core.md` with doctor/onboard updates
- [x] 8.10 Update `README.md` with all new commands and features

## 9. Build & Verify

- [x] 9.1 Run `go build ./...` — all packages compile
- [x] 9.2 Run `go test ./internal/cli/... ./internal/graph/... ./internal/agentmemory/...` — all tests pass
- [x] 9.3 Verify all new commands with `--help` output
- [x] 9.4 Run `/simplify` code review and fix all findings (8 issues fixed)
