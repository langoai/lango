# App Boundary Decoupling Design

## Context

`internal/app` has grown into a broad application package that owns process assembly, runtime wiring, tool construction, bridge adapters, and several CLI-facing helper surfaces. The package currently contains more than 15,000 non-test lines across 55 non-test Go files. This size is not the root problem by itself; the structural issue is that downstream CLI/TUI packages directly import `internal/app`, which makes future app decomposition expensive and risky.

The first architecture change will establish a dependency firewall before moving large wiring or tool files. The public CLI behavior remains stable, while internal APIs may change.

## Goals

- Make `cmd/lango` the only non-`internal/app` composition entrypoint that directly imports `internal/app`.
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

`internal/cli/cockpit` currently depends on `*app.StatusCollector` through `cockpit.Deps.FeatureStatuses`. Replace that concrete dependency with a small local interface owned by `internal/cli/cockpit`, because current usage is cockpit-specific.

The interface should expose only what the cockpit status page needs. `cmd/lango` can pass `*app.StatusCollector` into that interface slot because the concrete app type already has the required methods. Do not introduce a broader shared status package for this slice unless implementation discovers another production consumer.

### Status CLI Bridge

`internal/cli/status` currently creates a local app instance to access dead-letter tools through the app tool catalog. This is the riskiest part of the change because `deadLetterLoaderFromBoot` calls `app.New(boot, app.WithLocalChat())` and then reads `application.ToolCatalog`.

Resolve this by moving app-backed dead-letter bridge construction out of `internal/cli/status`. The status package should keep the existing `deadLetterBridge` and `deadLetterBridgeLoader` interfaces, but the loader that bootstraps an app and extracts a `*toolcatalog.Catalog` must live outside the CLI package. Prefer a small production bridge package, such as `internal/deadletterbridge`, that can build the catalog-backed bridge while `internal/cli/status` remains app-independent.

`internal/cli/status` should use the same catalog-backed bridge pattern already used by cockpit's `DeadLetterToolBridge`: command code depends on bridge behavior, not on app construction. `cmd/lango` wires the production loader into `NewStatusCmd`, while tests can keep using fake bridge loaders.

### Hook Registry Builder

`internal/cli/agent/hooks.go` uses `app.BuildHookRegistry(...)` to produce a registry snapshot. Move this construction function to `internal/toolchain`, because its signature already depends on `config`, `eventbus`, `toolchain.KnowledgeSaver`, and `toolcatalog`, and it returns `*toolchain.HookRegistry`.

`internal/app` can call the same `internal/toolchain` function after the move. The CLI agent command should depend on `internal/toolchain`, not on `internal/app`.

### Architecture Enforcement

Add an `internal/archtest` rule that scans production imports and fails when any package under `internal/cli/**` imports `github.com/langoai/lango/internal/app`.

The rule must use exact import path matching with a slash boundary, following the existing pattern:

```go
importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
```

This avoids false positives for neighboring packages such as `internal/approval` or `internal/appinit`. The rule should follow the repository's existing archtest pattern: use `go list -json` via `os/exec`, inspect the `Imports` field, and avoid adding external dependencies. Because `TestImports` is not inspected, test-only imports are excluded by design.

## Data Flow

Runtime assembly remains:

```text
cmd/lango -> bootstrap -> app.New(...) -> assembled app/runtime dependencies
```

CLI/TUI consumption becomes:

```text
cmd/lango -> internal/cli/* constructors -> narrow interfaces / DTOs / app-independent helpers
```

CLI packages should not inspect app internals or depend on `app.App` fields. They should receive the exact dependency needed for rendering, command execution, or bridge access.

## Error Handling

The change must preserve current user-facing error behavior:

- Bootstrap failures still return actionable `bootstrap: ...` style errors where they do today.
- Missing tool catalog or missing dead-letter tools still produce explicit unavailable errors.
- Cockpit optional dependencies remain optional and should render disabled or placeholder states rather than panic.
- Hook registry snapshot output should preserve existing JSON and text fields.

## Testing

Verification requires:

- Update or add unit tests for affected CLI packages using fakes for the new ports.
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

- `rg 'github.com/langoai/lango/internal/app(\"|/)' internal/cli --glob '*.go'` returns no production CLI imports.
- `cmd/lango` remains allowed to import and instantiate `internal/app`.
- A production import from `internal/cli/**` to `internal/app` fails the architecture test.
- Existing status, cockpit, and agent hooks behavior remains compatible at the CLI contract level.
- `go build ./...` passes.
- `go test ./...` passes, or any pre-existing environment-dependent failures are explicitly reproduced and documented before implementation begins.
