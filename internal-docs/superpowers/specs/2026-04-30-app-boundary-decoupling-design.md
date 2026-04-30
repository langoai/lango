# App Boundary Decoupling Design

## Context

`internal/app` has grown into a broad application package that owns process assembly, runtime wiring, tool construction, bridge adapters, and several CLI-facing helper surfaces. For orientation, the package currently contains more than 15,000 non-test lines across 55 non-test Go files. The size is not the root problem by itself; the structural issue is that downstream CLI/TUI packages directly import `internal/app`, which makes future app decomposition expensive and risky.

The first architecture change will establish a dependency firewall before moving large wiring or tool files. The public CLI behavior remains stable, while internal APIs may change.

## Goals

- Make `cmd/lango` the only production importer of `internal/app`.
- Remove direct `internal/app` imports from `internal/cli/**`.
- Replace CLI/TUI dependencies on concrete app types and helper functions with narrow interfaces, DTOs, or app-independent packages.
- Add an architecture test that prevents `internal/cli/**` from importing `internal/app` again.
- Preserve existing command names, flags, output schemas, and user-facing error behavior.

## Non-Goals

- Do not split `internal/app/tools_meta.go` in this change.
- Do not move P2P, economy, workspace, or MCP wiring packages in this change.
- Do not rewrite `app.New()` or change the application startup order.
- Do not change public CLI command names, flags, or output schema.
- Do not perform a broad README or public documentation rewrite.

## Target Boundary

```text
cmd/lango
  -> internal/app        // allowed: process-level composition root
  -> internal/cli/*      // allowed: user-facing commands and TUI

internal/cli/*
  -> internal/app        // forbidden
  -> narrow runtime ports, DTOs, or app-independent helper packages

internal/app
  -> internal/appinit, domains, tools, gateway, channels
```

`cmd/lango` may continue to call `app.New(...)`, `app.WithLocalChat()`, and other process assembly options. CLI and TUI packages must receive already-assembled dependencies or app-independent construction functions.

## Components

### Runtime Status Port

`internal/cli/cockpit` currently depends on `*app.StatusCollector` through `cockpit.Deps.FeatureStatuses`, but the dependency is not used by `cockpit.New`. The actual status page wiring in `cmd/lango` already passes `application.FeatureStatuses.All` as a `func() []types.FeatureStatus` to `pages.NewStatusPage`.

Remove `cockpit.Deps.FeatureStatuses` instead of adding a new interface. Keep the status page contract as `func() []types.FeatureStatus`, and update `cmd/lango` in both wiring points:

- Remove `FeatureStatuses: application.FeatureStatuses` from the `cockpit.Deps` literal.
- Keep the existing `statusProvider = application.FeatureStatuses.All` method value for `pages.NewStatusPage`.

`internal/gateway.Server.SetFeatureStatuses(statuses []types.FeatureStatus)` already uses a slice contract and does not need to change.

### Status CLI Bridge

`internal/cli/status` currently creates a local app instance to access dead-letter tools through the app tool catalog. This is the riskiest part of the change because `deadLetterLoaderFromBoot` calls `app.New(boot, app.WithLocalChat())` and then reads `application.ToolCatalog`.

Resolve this by moving app-backed dead-letter bridge construction out of `internal/cli/status` and into `cmd/lango`. Do not create another production package that imports `internal/app`; that would weaken the composition-root goal. `cmd/lango` should own a lazy helper that bootstraps the app and passes a catalog-backed bridge loader into the status command.

`internal/cli/status` should use the same catalog-backed bridge pattern already used by cockpit's `DeadLetterToolBridge`: command code depends on bridge behavior, not on app construction. `NewStatusCmd` should accept the production dead-letter loader as a dependency, while tests can keep using fake bridge loaders. App-backed construction belongs in a new `cmd/lango/dead_letter_status.go` file rather than further expanding `cmd/lango/main.go`.

Export exactly the status package API needed for that wiring:

- `type DeadLetterBridge interface`
- `type DeadLetterBridgeLoader func() (DeadLetterBridge, func(), error)`
- `type DeadLetterListOptions struct`
- `type DeadLetterListPage struct`
- `func NewToolCatalogDeadLetterBridge(catalog *toolcatalog.Catalog) DeadLetterBridge`
- `func NewStatusCmd(bootLoader func() (*bootstrap.Result, error), deadLetterLoader DeadLetterBridgeLoader) *cobra.Command`

`NewStatusCmd` should treat a nil `deadLetterLoader` as "dead-letter status tools are not available" for dead-letter subcommands, so tests and future callers can construct the base status command without importing app wiring. Keep retry result, summary result, trend, and rendering helper types unexported unless an implementation test proves `cmd/lango` needs them. `internal/cli/status` has one app-backed entry point today: `deadLetterLoaderFromBoot`. Moving that single loader is enough to make the status package app-free; the base `status` command and all dead-letter subcommands continue to share the same loader dependency.

This preserves the current lazy bootstrap behavior for dead-letter subcommands, but it does not make that bootstrap lighter. Today the path builds a full app only to access three dead-letter tool handlers. Reducing that startup cost is a follow-up change, not part of this boundary slice.

Do not merge the cockpit dead-letter option types with status in this slice. Cockpit, status, and `postadjudicationstatus` currently use related but differently shaped option structs. Consolidating them is valid follow-up cleanup, but doing it here would expand the boundary change into a dead-letter API refactor.

### Hook Registry Builder

`internal/cli/agent/hooks.go` uses `app.BuildHookRegistry(...)` to produce a registry snapshot. Move this construction function to `internal/toolchain/build_hook_registry.go` as exported `toolchain.BuildHookRegistry`, because its signature already depends on `config`, `eventbus`, `toolchain.KnowledgeSaver`, and `toolcatalog`, and it returns `*toolchain.HookRegistry`.

`internal/app` should call the same `internal/toolchain` function after the move. The CLI agent command should depend on `internal/toolchain`, not on `internal/app`.

Implementation inventory:

- Update `internal/cli/agent/hooks.go` to call `toolchain.BuildHookRegistry`.
- Move `internal/app/build_hook_registry_test.go` to the `internal/toolchain` package.
- Update `internal/app/app.go` so the runtime bootstrap calls `toolchain.BuildHookRegistry` directly.
- Delete the private `buildHookRegistry` wrapper in `internal/app/app.go`; keeping it would be a dead abstraction layer.
- Verify `internal/config`, `internal/eventbus`, and `internal/toolcatalog` do not import `internal/toolchain`; current code shows no such reverse import, so the move should not create a package cycle.

### Architecture Enforcement

Add an `internal/archtest` rule that scans production imports and fails when any package under `internal/cli/**` imports `github.com/langoai/lango/internal/app`.

The rule must use exact import path matching with a slash boundary, following the existing pattern:

```go
importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
```

The prefix must be the exact module-qualified app import path: `github.com/langoai/lango/internal/app`. This avoids false positives for neighboring packages such as `internal/approval` or `internal/appinit`. The rule should follow the repository's existing archtest pattern: use `go list -json` via `os/exec`, inspect the `Imports` field, and avoid adding external dependencies. Because `TestImports` is not inspected, test-only imports are excluded by design.

## OpenSpec Contract Impact

Before implementation, inspect these specs and create delta updates when concrete package contracts are affected:

- `openspec/specs/feature-status/spec.md`
- `openspec/specs/cockpit-status-page/spec.md`
- `openspec/specs/cli-agent-tools-hooks/spec.md`
- `openspec/specs/application-core/spec.md`

Known expected deltas:

- `cli-agent-tools-hooks` currently requires `internal/app` to export `BuildHookRegistry`; this must change to `internal/toolchain`.
- `cockpit-status-page` currently references `FeatureStatuses.All()`; it should describe the `func() []types.FeatureStatus` status provider used by the page instead of app concrete type knowledge.
- `feature-status` may keep `StatusCollector` in the app layer for wiring aggregation, but CLI/TUI consumption must be described through `types.FeatureStatus` slices or provider functions rather than app imports.
- `application-core` should remain focused on centralized process assembly and must not require CLI packages to import `internal/app`.

## Data Flow

Runtime assembly remains:

```text
cmd/lango -> bootstrap -> app.New(...) -> assembled app/runtime dependencies
```

CLI/TUI consumption becomes:

```text
cmd/lango -> internal/cli/* constructors -> narrow interfaces / function providers / DTOs / app-independent helpers
```

CLI packages should not inspect app internals or depend on `app.App` fields. They should receive the exact dependency needed for rendering, command execution, or bridge access.

## Implementation Sequence

Apply the implementation in this order so each checkpoint has a clear rollback boundary and the architecture test is only enabled after the code can pass it:

1. Move `BuildHookRegistry` to `internal/toolchain/build_hook_registry.go`, update app and agent callers, and move the tests.
2. Remove the unused `cockpit.Deps.FeatureStatuses` field and update `cmd/lango` cockpit wiring.
3. Export the status dead-letter bridge API, move app-backed lazy loader construction to `cmd/lango/dead_letter_status.go`, and update status tests.
4. Add and enable the `internal/archtest` rule that blocks production `internal/cli/** -> internal/app` imports.

## Error Handling

The change must preserve current user-facing error behavior:

- Bootstrap failures still return actionable `bootstrap: ...` style errors where they do today.
- Missing tool catalog or missing dead-letter tools still produce explicit unavailable errors.
- Cockpit optional dependencies remain optional and should render disabled or placeholder states rather than panic.
- Hook registry snapshot output should preserve existing JSON and text fields.

## Testing

Verification requires:

- Update or add unit tests for affected CLI packages using fakes for the new ports.
- Verify `lango agent hooks --json` remains decoded-JSON field compatible for a fixed config fixture before and after the `BuildHookRegistry` move. Do not require byte-for-byte equivalence unless saveable tool ordering is explicitly stabilized.
- Add or update `internal/archtest` coverage for the CLI-to-app import boundary.
- Run `go test ./internal/archtest/...` during implementation.
- Run `go build ./...` and `go test ./...` before reporting completion.

## Teammate Review Perspectives

- Architect: verify the dependency direction and ensure `internal/app` remains a composition root rather than a shared utility package.
- Core Developer: ensure new ports are small, stable, and do not leak concrete app types.
- UI/UX Developer: verify command behavior, flags, output, and cockpit behavior remain unchanged.
- QA / Tester: verify negative paths for missing runtime dependencies, bootstrap failures, and unavailable tools.
- Technical Writer: avoid public documentation changes unless behavior changes; keep architecture notes internal for this slice.

## OpenSpec Execution Plan

Implementation should use an OpenSpec change focused on app boundary decoupling. The change should follow the repository workflow: create or fast-forward artifacts, apply implementation tasks, verify against the artifacts, sync specs if needed, and archive after successful build and test verification.

## Acceptance Criteria

- `rg 'github.com/langoai/lango/internal/app["/]' internal/cli --glob '*.go'` returns no production CLI imports.
- Baseline before this change: `rg 'github.com/langoai/lango/internal/app["/]' --glob '*.go'` returns four production files: `cmd/lango/main.go`, `internal/cli/cockpit/deps.go`, `internal/cli/status/status.go`, and `internal/cli/agent/hooks.go`.
- After this change: the same production grep returns only `cmd/lango/main.go`.
- `cmd/lango` remains allowed to import and instantiate `internal/app`.
- A production import from `internal/cli/**` to `internal/app` fails the architecture test.
- OpenSpec deltas are added for affected concrete package contracts before implementation tasks are marked complete.
- Existing status, cockpit, and agent hooks behavior remains compatible at the CLI contract level.
- `lango agent hooks --json` preserves the same decoded JSON fields for a fixed config fixture.
- `go build ./...` passes.
- `go test ./...` passes, or any pre-existing environment-dependent failures are explicitly reproduced and documented before implementation begins.
