## 1. Stage 1 — NormalizePaths sandbox expansion (deferred fix from PR 3)

- [x] 1.1 Add three sandbox path normalizations to `internal/config/loader.go:NormalizePaths` (`WorkspacePath`, `OS.SeatbeltCustomProfile`, `AllowedWritePaths` slice)
- [x] 1.2 Add new helper `normalizePathSlice([]string, dataRoot, home) []string` to `internal/config/loader.go` (allocates new slice, preserves empty entries)
- [x] 1.3 Update doc comments on `WorkspacePath`, `AllowedWritePaths`, `OS.SeatbeltCustomProfile` in `internal/config/types_sandbox.go` to note "normalized in PostLoad"
- [x] 1.4 Add `TestNormalizePaths_Sandbox` in `internal/config/loader_test.go` with 8 subtests (tilde, relative, empty, absolute, slice, nil slice, SeatbeltCustomProfile, idempotency)
- [x] 1.5 Run `go build ./...`, `GOOS=linux GOARCH=amd64 go build ./...`, `go test ./...`, `golangci-lint run ./...` — all pass
- [x] 1.6 Stop at commit boundary; user commits manually

## 2. Stage 2 — Policy signature + control-plane denylist + 3 sandbox sites dataRoot injection

- [x] 2.1 Run `Grep "DefaultToolPolicy|StrictToolPolicy|MCPServerPolicy"` to inventory all call sites before editing
- [x] 2.2 Change three policy helper signatures in `internal/sandbox/os/policy.go`: `DefaultToolPolicy(workDir, dataRoot)`, `StrictToolPolicy(workDir, dataRoot)`, `MCPServerPolicy(dataRoot)`
- [x] 2.3 Move `.git` denial from `StrictToolPolicy` into the baseline `DefaultToolPolicy`; collapse `StrictToolPolicy` into a wrapper around `DefaultToolPolicy`
- [x] 2.4 Add `dataRoot` to `DenyPaths` in all three helpers when the argument is non-empty (resolved via `filepath.Abs`)
- [x] 2.5 Update `internal/sandbox/os/policy_test.go` with new cases: `TestDefaultToolPolicy_EmptyDataRoot`, `TestMCPServerPolicy_EmptyDataRoot`, plus signature updates for existing cases
- [x] 2.6 Add Seatbelt profile generation case verifying `(deny file-write* (subpath "<dataRoot>"))` is emitted
- [x] 2.7 Update `internal/sandbox/os/bwrap_args_test.go`: signature updates + new `TestCompileBwrapArgs_DenyOverlapsWritePath` (deny mount comes after write mount so last-mount-wins yields deny precedence)
- [x] 2.8 Update `internal/sandbox/os/bwrap_linux_test.go`: import `os`, mkdir `<work>/.git` (now baseline-denied), pass empty dataRoot
- [x] 2.9 Update `internal/supervisor/supervisor.go`: call `DefaultToolPolicy(workDir, cfg.DataRoot)`, append `cfg.Sandbox.AllowedWritePaths` to `policy.Filesystem.WritePaths`
- [x] 2.10 Add `dataRoot` field to `internal/skill/executor.go:Executor`; change `SetOSIsolator` signature to `(iso, workspacePath, dataRoot)`; update `executeScript` apply to pass `e.dataRoot`
- [x] 2.11 Update `internal/skill/registry.go:Registry.SetOSIsolator` pass-through to forward `dataRoot`
- [x] 2.12 Update `internal/skill/executor_test.go` and `internal/skill/registry_test.go` (if any) call sites for new signature
- [x] 2.13 Add `dataRoot` field to `internal/mcp/connection.go:ServerConnection`; change `SetOSIsolator(iso, dataRoot)`; update `createTransport` to call `MCPServerPolicy(sc.dataRoot)`
- [x] 2.14 Add `dataRoot` field to `internal/mcp/manager.go:ServerManager`; change `SetOSIsolator(iso, dataRoot)`; propagate to all existing and future connections in `ConnectAll`
- [x] 2.15 Update `internal/mcp/connection_test.go` call sites for the new signature
- [x] 2.16 Update `internal/app/wiring_knowledge.go:229` to call `registry.SetOSIsolator(iso, workDir, cfg.DataRoot)`
- [x] 2.17 Update `internal/app/wiring_mcp.go:42` to call `mgr.SetOSIsolator(iso, cfg.DataRoot)`
- [x] 2.18 Run full verification: `go build ./...`, `GOOS=linux GOARCH=amd64 go build ./...`, `go test ./...`, `golangci-lint run ./...`, `go run ./cmd/lango sandbox test` (4/4 PASS)
- [x] 2.19 Re-run the call site inventory grep at end of stage to verify zero misses
- [x] 2.20 Stop at commit boundary; user commits manually

## 3. Stage 3 — ExcludedCommands + SandboxDecisionEvent + audit + 3 publish sites

- [x] 3.1 Add `"sandbox_decision"` value to the `audit_log.action` enum in `internal/ent/schema/audit_log.go`
- [x] 3.2 Run `go generate ./internal/ent` to regenerate `internal/ent/auditlog/auditlog.go` and related files
- [x] 3.3 Add `EventSandboxDecision = "sandbox.decision"` constant + `SandboxDecisionEvent` struct + `EventName()` method + `PublishSandboxDecision` helper to `internal/eventbus/events.go`
- [x] 3.4 Add `r.handleSandboxDecision` handler + `SubscribeTyped[SandboxDecisionEvent]` registration to `internal/observability/audit/recorder.go` (conditional `SetSessionKey` for empty session)
- [x] 3.5 Add `Sandbox.ExcludedCommands []string` field to `internal/config/types_sandbox.go` with sh-trap-aware doc comment
- [x] 3.6 Add `path/filepath`, `eventbus`, `session` imports to `internal/tools/exec/exec.go`
- [x] 3.7 Add `ExcludedCommands []string`, `Bus *eventbus.Bus` fields to `exec.Config`; add `fallbackOnce sync.Once` field to `exec.Tool`
- [x] 3.8 Add `Tool.SetEventBus(bus)` setter
- [x] 3.9 Change `applySandbox` signature to `applySandbox(ctx, cmd, userCommand string)`; implement excluded check, all 4 decision branches with publish, fallback warning
- [x] 3.10 Add helpers: `excludedMatch` (first-token basename), `publishDecision` (ctx-derived sessionKey), `warnFallbackOnce` (sync.Once)
- [x] 3.11 Update three call sites in exec.go (`Run`, `RunWithPTY`, `StartBackground`) to pass raw `command` to `applySandbox`
- [x] 3.12 Add `eventbus` import to `internal/supervisor/supervisor.go`; populate `execConfig.ExcludedCommands` from cfg; add `Supervisor.SetEventBus(bus)` forwarder method
- [x] 3.13 Add B1a in `internal/app/app.go` after `populateAppFields`: resolve foundation values and call `Supervisor.SetEventBus(bus)`
- [x] 3.14 Add `eventbus`, `session` imports to `internal/skill/executor.go`; add `bus` field + `SetEventBus` setter + `publishSandboxDecision` helper
- [x] 3.15 Add publish calls in `executeScript` for all 4 decision branches (applied / skipped / rejected w/ isolator, rejected/skipped w/o isolator)
- [x] 3.16 Add `eventbus` import to `internal/skill/registry.go`; add `Registry.SetEventBus(bus)` pass-through
- [x] 3.17 Add `bus *eventbus.Bus` parameter to `initSkills(cfg, baseTools, bus)` in `internal/app/wiring_knowledge.go`; call `registry.SetEventBus(bus)`
- [x] 3.18 Update `internal/app/modules.go:312` to pass `m.bus` to `initSkills`
- [x] 3.19 Add `eventbus` import to `internal/mcp/connection.go`; add `bus` field + `SetEventBus` setter + `publishSandboxDecision` helper
- [x] 3.20 Add publish calls in `createTransport` for all 4 decision branches; SessionKey intentionally empty (process-level)
- [x] 3.21 Add `eventbus` import to `internal/mcp/manager.go`; add `bus` field + `SetEventBus` setter; propagate via `ConnectAll`
- [x] 3.22 Add `bus *eventbus.Bus` parameter to `initMCP(cfg, bus)` in `internal/app/wiring_mcp.go`; call `mgr.SetEventBus(bus)`
- [x] 3.23 Update `internal/app/modules.go:927` to pass `m.bus` to `initMCP`
- [x] 3.24 Add `TestExcludedMatch` (7 cases) to `internal/tools/exec/exec_test.go`
- [x] 3.25 Add `TestApplySandbox_ExcludedDoesNotMatchSh` regression guard to `internal/tools/exec/exec_test.go` (pins the sh-wrapping semantic)
- [x] 3.26 Add `TestApplySandbox_ExcludedBypass` to `internal/tools/exec/exec_test.go`
- [x] 3.27 Run `Grep "isolator.Apply"` inventory at end of stage to verify all three sites publish
- [x] 3.28 Run full verification: build (native + cross), test, lint, smoke test
- [x] 3.29 Stop at commit boundary; user commits manually

## 4. Stage 4 — `lango sandbox status` Recent Decisions section + TUI

- [x] 4.1 Add `entgo.io/ent/dialect/sql`, `bootstrap`, `ent/auditlog` imports to `internal/cli/sandbox/sandbox.go`
- [x] 4.2 Add `BootLoader func() (*bootstrap.Result, error)` type alias
- [x] 4.3 Change `NewSandboxCmd(cfgLoader, bootLoader)` and `newStatusCmd(cfgLoader, bootLoader)` signatures
- [x] 4.4 Add `--session <prefix>` flag to status command
- [x] 4.5 Add `renderRecentDecisions(ctx, w, bootLoader, sessionPrefix)` helper with graceful degradation (nil loader, loader error, nil DBClient all silently skip)
- [x] 4.6 Add `truncateSessionKey(key, width)` helper that pads empty keys with dashes and truncates long keys
- [x] 4.7 Update `cmd/lango/main.go:213` to pass `cliboot.BootResult` to `NewSandboxCmd`
- [x] 4.8 Add `os_sandbox_excluded_commands` InputText field to `internal/cli/settings/forms_sandbox.go` with UNSANDBOXED warning in description
- [x] 4.9 Add `os_sandbox_excluded_commands` case to `internal/cli/tuicore/state_update.go` mapping to `cfg.Sandbox.ExcludedCommands` via `splitCSV`
- [x] 4.10 Create `internal/cli/sandbox/sandbox_test.go` with `TestTruncateSessionKey` (4 cases) and 3 graceful-degradation tests for `renderRecentDecisions`
- [x] 4.11 Run full verification: build (native + cross), test, lint, smoke test
- [x] 4.12 Stop at commit boundary; user commits manually

## 5. Stage 5 — Documentation downstream sync

- [x] 5.1 Update `README.md` feature line + sandbox config table (3 new rows: workspacePath, allowedWritePaths, excludedCommands)
- [x] 5.2 Update `docs/configuration.md`: add control-plane masking + fail-open visibility paragraphs, JSON example with `excludedCommands`, table rows for new fields
- [x] 5.3 Update `docs/cli/sandbox.md`: add Recent Sandbox Decisions section with example output, ExcludedCommands semantics paragraph, fail-open warning paragraph, `--session` flag table row
- [x] 5.4 Update `prompts/TOOL_USAGE.md`: add OS sandbox awareness bullet under Exec Tool covering control-plane denial, ExcludedCommands semantics, and "do not invent shell tricks to bypass"
- [x] 5.5 Update `prompts/SAFETY.md`: add "Control-plane is off-limits" bullet naming config / DB / secret tokens / skills as denied surfaces
- [x] 5.6 Run final verification: build, test, lint
- [x] 5.7 Stop at commit boundary; user commits manually

## 6. Stage 6 — OpenSpec change + verify + sync + archive

- [x] 6.1 Run `openspec new change sandbox-escape-hardening`
- [x] 6.2 Write `proposal.md` (Why / What Changes / Capabilities / Impact)
- [x] 6.3 Write `design.md` (Context / Goals / D1~D11 / Risks / Migration Plan)
- [x] 6.4 Write delta spec `specs/os-sandbox-core/spec.md` (MODIFIED Policy types + Seatbelt generation)
- [x] 6.5 Write delta spec `specs/os-sandbox-cli/spec.md` (ADDED Recent Decisions section + TUI excluded commands field)
- [x] 6.6 Write delta spec `specs/os-sandbox-integration/spec.md` (MODIFIED exec/skill/MCP integration with publish + dataRoot wiring)
- [x] 6.7 Write delta spec `specs/mcp-integration/spec.md` (MODIFIED MCP stdio sandbox to take dataRoot + publish)
- [x] 6.8 Write new capability spec `specs/sandbox-exception-policy/spec.md` (5 requirements: ExcludedCommands match / SandboxDecisionEvent schema / audit subscribe / fail-open warn / 3 publish sites + config field)
- [x] 6.9 Write `tasks.md` (this file)
- [x] 6.10 Run `/opsx:verify` and address any CRITICAL findings
- [x] 6.11 Sync delta specs into `openspec/specs/` (apply MODIFIED requirement replacements + ADDED requirements + new capability)
- [x] 6.12 Run `/opsx:archive` to move the change directory into `openspec/changes/archive/YYYY-MM-DD-sandbox-escape-hardening/`
- [ ] 6.13 Stop at commit boundary; user commits the sync+archive bundle manually
