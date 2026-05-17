## Context

The extension command family intentionally has documented exit codes beyond the root command's generic `1`: `1` for user-facing extension errors, `2` for internal/usage errors, and `3` for user-declined or non-interactive confirmation failures. The current implementation preserves those codes by calling a package-level `os.Exit` seam from `internal/cli/extension`, but that pushes process ownership below `cmd/lango/main.go`.

## Goals

- Keep all process termination in binary entrypoints.
- Preserve extension command exit-code semantics exactly.
- Remove panic-based test seams from extension command tests.
- Add a narrow reusable exit-code error helper for future CLI packages that need non-generic exit codes.

## Non-Goals

- Do not change extension command UX, flags, or output formats.
- Do not redesign Cobra error rendering across all commands.
- Do not broaden this change into a full CLI framework refactor.

## Decisions

### Shared exit-code error helper

Add a small `internal/cli/cliexit` package that exposes:

- `New(code int, err error) error`
- `Code(err error) (int, bool)`
- an error type that supports `errors.As` and `errors.Unwrap`

This avoids import cycles and keeps exit-code semantics reusable without coupling `internal/cli/extension` to `cmd/lango`.

### Root-owned process exit

`cmd/lango/runMain` will inspect errors returned by `rootCmd.Execute()`. If the error carries a `cliexit` code, `runMain` prints the error message once and returns that code. Otherwise it preserves the existing generic `1` behavior.

### Extension command behavior

`internal/cli/extension.cliExit` will set `SilenceUsage` and `SilenceErrors`, then return a `cliexit` error instead of printing and terminating. Direct command tests will assert `cliexit.Code(err)` rather than intercepting panic.

### Architecture guard

Extend repository-level tests to scan non-test Go files under `internal/cli` and fail on direct `os.Exit` references. The new guard prevents the same pattern from reappearing in another CLI package.

## Risks

- Cobra may print errors if `SilenceErrors` is not set before returning `cliexit`; extension `cliExit` must keep setting it.
- If both extension `cliExit` and `runMain` print the same error, users would see duplicate output; therefore `cliExit` should stop printing.
- Tests that previously expected panic-based process exits must be updated to inspect returned errors.
