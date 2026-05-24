## Why

`internal/cli/extension` currently calls `os.Exit` from inside the command package, bypassing the root command's normal error and exit-code path. This conflicts with the documented project convention that process termination is owned by `cmd/*/main.go`, makes direct command tests rely on panic seams, and weakens architecture enforcement.

## What Changes

- Replace extension CLI process exits with structured errors carrying the intended exit code.
- Teach the `lango` entrypoint to preserve those structured exit codes when Cobra returns them.
- Add a repository guard that rejects direct `os.Exit` usage in non-test `internal/cli` packages.
- Keep the documented extension exit codes unchanged: user errors exit 1, internal errors exit 2, and user-declined confirmations exit 3.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `extension-pack-cli`: Preserve extension CLI exit-code semantics without process exits inside `internal/cli/extension`.
- `production-readiness`: Enforce that internal CLI packages do not terminate the process directly.
- `test-coverage`: Add executable coverage for structured CLI exit-code propagation and internal CLI `os.Exit` hygiene.

## Impact

- Affected code: `internal/cli/extension`, `cmd/lango`, and repository-level quality tests.
- Affected tests: extension command tests, main exit-code tests, and production quality guard tests.
- No public command syntax changes and no breaking change to documented CLI exit codes.
