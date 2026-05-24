## Overview

This change is intentionally narrow. `newKMSKeysCmd` already centralizes all user-facing rendering for `lango security kms keys`, so the implementation only needs to replace direct stdout writes with the Cobra command writer.

## Decisions

### Use Cobra output stream for all non-error output

- JSON output uses `json.NewEncoder(cmd.OutOrStdout())`
- Empty-state text uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`
- Tabular text output uses `fmt.Fprintf(cmd.OutOrStdout(), ...)`

This keeps the command aligned with other modernized CLI surfaces and allows tests to verify output without swapping process-global stdout.

### Reuse persistent KeyRegistry-backed test seam

Existing `keyRegistryBootLoader(...)` test wiring already provides a stable way to seed persistent registry entries. The new tests should continue using that seam and assert against command-writer output only.

## Non-Goals

- No changes to KMS key data shape or sorting behavior
- No changes to `kms status`, `kms test`, `kms wrap`, or `kms detach`
