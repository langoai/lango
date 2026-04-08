## Why

A second Codex review pass, run after `sandbox-escape-hardening-fixes` was archived, surfaced two additional regressions that PR 4 (and its first-round fix) left unresolved. Both are small but independently worth fixing before the next release.

1. **P1 — bwrap mount order.** `compileBwrapArgs` emits `--proc /proc`, `--dev /dev`, `--tmpfs /run` BEFORE `--ro-bind / /`. bubblewrap processes options left-to-right, so the later root bind shadows the earlier specialised mounts and sandboxed children end up seeing the host's `/proc` and `/dev`. This weakens PID namespace isolation and device filtering. The bug predates PR 4 (introduced in PR 3 when the bwrap backend first landed), but it is small and fits in the same testing scope as the other sandbox hardening fixes.
2. **P2 — Relative sandbox paths collide with the DataRoot deny.** `NormalizePaths` resolves relative `sandbox.workspacePath` / `sandbox.allowedWritePaths` entries under `cfg.DataRoot`. `DefaultToolPolicy` then adds `cfg.DataRoot` to `DenyPaths`. At runtime the deny tmpfs (bwrap) or deny rule (Seatbelt) covers the user's workspace, silently breaking writes. Narrow impact (only users with relative sandbox paths) but mysterious failure mode.

Two further round-2 findings — subdirectory walk-up `.git` discovery and MCP workspace `.git` deny — are intentionally deferred to PR 5 because they expand PR 4's contract rather than restore it.

## What Changes

- **`internal/sandbox/os/bwrap_args.go`**: Reorder `compileBwrapArgs` so `--ro-bind / /` appears BEFORE `--proc /proc`, `--dev /dev`, and `--tmpfs /run`. The specialised mounts are now layered on top of the root bind, matching the canonical bwrap wrapper pattern.
- **`internal/sandbox/os/bwrap_args_test.go`**: New `TestCompileBwrapArgs_RootBindBeforeSpecialMounts` regression guard that asserts the argv index order directly.
- **`internal/config/loader.go`**: New `pathIsUnder(child, parent string) bool` helper. `Validate` rejects `sandbox.workspacePath` and every entry of `sandbox.allowedWritePaths` that resolves to `cfg.DataRoot` itself or to a subtree of it, with an actionable error message naming the colliding path.
- **`internal/config/loader_test.go`**: Extended `TestValidate` with five new subtests covering rejection/acceptance around DataRoot overlap. New `TestPathIsUnder` with eight cases covering nested, same, sibling, parent-is-child, trailing separator, and empty input paths.

## Capabilities

### New Capabilities

(none — pure fix change)

### Modified Capabilities

- `os-sandbox-core`: Add two new requirements — `bwrap mount ordering` (argv index order assertion) and `Sandbox path validation against DataRoot overlap` (config startup rejection).

## Impact

- **Affected code**: `internal/sandbox/os/bwrap_args.go`, `internal/sandbox/os/bwrap_args_test.go`, `internal/config/loader.go`, `internal/config/loader_test.go`.
- **Affected specs**: `os-sandbox-core` (two new requirements added; no existing requirements modified — the Round 1 modifications were already synced when `sandbox-escape-hardening-fixes` archived).
- **Documentation**: No README / docs / prompts changes. The runtime-visible impact is narrow and the fixes restore expected behaviour rather than introduce new user-facing contracts.
- **Runtime behavior**:
  - Linux: sandboxed children now see a fresh `/proc` (new PID namespace), filtered `/dev`, and an empty `/run`. Previously the host's `/proc` leaked in.
  - Config: invalid sandbox paths under `cfg.DataRoot` are rejected at startup with a clear error instead of silently breaking workspace writes.
- **Out of scope (PR 5)**: subdirectory walk-up `.git` discovery; MCP workspace `.git` deny; `bwrap --version` probe false positive on hosts with unprivileged userns disabled; native Landlock+seccomp backend; file-level deny; symlink chain resolution; glob/path semantics normalization; per-tool policy override.
