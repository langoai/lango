## Context

`lango agent hooks` is a config-only command (`cli/agent/hooks.go`) using `cfgLoader`. It outputs the 5 `HooksConfig` boolean flags and `BlockedCommands`. The actual `HookRegistry` is built inside `app.go:buildHookRegistry(cfg, bus)` — an unexported function that conditionally registers SecurityFilter, AccessControl, and EventBus hooks based on config. `KnowledgeSaveHook` is constructed separately via the learning/knowledge wiring and currently receives an **empty** `SaveableTools` set (no default list is defined).

Provenance already snapshots the registry at app boot via `buildHookRegistrySnapshot()` in `modules_provenance.go`, demonstrating that hook name + priority extraction works. However, this snapshot is only written to provenance metadata, not surfaced in CLI output.

## Goals / Non-Goals

**Goals:**
- `lango agent hooks` displays registered pre/post hooks (name, priority, phase) in addition to config flags
- `KnowledgeSaveHook.SaveableTools` has a sensible app-level default constant so the hook actually saves read-type tool results
- Existing JSON/text output is backward compatible (additive fields only)
- The CLI does NOT require a full app bootstrap — a lightweight `BuildHookRegistry(cfg)` public helper is sufficient

**Non-Goals:**
- Adding a `SaveableTools` config field (P1 — catalog capability-based)
- Surfacing child session lifecycle observers (separate concern, wrong CLI scope)
- Modifying `doctor` checks (Unit A/C territory)
- User-defined hook registration API

## Decisions

### D1: Public helper `BuildHookRegistry(cfg) → *HookRegistry`

**Choice**: Extract the body of `app.go:buildHookRegistry` into a public function. The private function calls the public one.

**Why not inject the live `HookRegistry` from a running app?**
Running a full bootstrap to open `hooks.go` would require DB + crypto init (5-phase pipeline). The CLI command should stay fast and dependency-free. Since the registry is deterministically derived from config, rebuilding it from config alone produces an honest snapshot.

**Caveat**: The EventBus hook branch checks `bus != nil`. In CLI mode there is no bus, so EventBus hook will be absent from the snapshot if the config enables it. We accept this and document it: "Snapshot is config-derived; hooks requiring a running bus show as `not wirable` in CLI mode."

### D2: Default `SaveableTools` as package constant

**Choice**: `toolchain.DefaultSaveableTools = []string{...}` hard-coded list of read-type tool names. Used by `app.go` builder as the default when wiring `KnowledgeSaveHook`.

**Why not catalog capability?**
Catalog capability extension (adding a `KnowledgeSaveable` field to tool metadata) is a metamodel change affecting all tool registrations. Deferred to P1. A constant is minimal, correct, and easy to audit.

**Tool name selection**: Inspect actual tool names registered in `internal/agent/tools_*.go` and `internal/toolcatalog/` before picking the list. Only read-side tools (e.g. `read_file`, `search`, `grep_search`, `list_files`) — never write-side tools.

### D3: Additive JSON schema

**Choice**: Keep all existing JSON fields. Add a new `registry` object with `{preHooks: [{name, priority, wirable}], postHooks: [{name, priority, wirable, details}]}`. `details` is an optional object (e.g. for KnowledgeSaveHook, contains `saveableTools: [...]`).

### D4: No TUI settings form change

The `SaveableTools` default is a code constant, not a config field. `forms_hooks.go` requires no modification since no new config field is introduced.

## Risks / Trade-offs

- **[Risk] EventBus hook absent in CLI snapshot** → Documented in output as `wirable: false` with reason. Acceptable because `doctor ToolHooksCheck` already covers runtime health.
- **[Risk] Tool names drift** → The hard-coded allowlist may become stale if tools are renamed. Mitigation: unit test that asserts all `DefaultSaveableTools` entries exist in the tool catalog at build time (or integration test).
- **[Risk] `buildHookRegistry` duplication** → Mitigated by having the private func call the public one. Single source of truth.
