## Why

`lango security kms keys` still writes operator-facing output directly to process stdout. That bypasses Cobra command writers and makes wrappers, tests, and embedding harnesses unable to capture text and JSON output consistently.

## What Changes

- Route `lango security kms keys` table, empty-state, and JSON output through `cmd.OutOrStdout()`
- Add command-level regression tests that capture output from the Cobra command writer
- Update public docs and OpenSpec to document the writer-routing contract

## Impact

- Makes `lango security kms keys` consistent with the rest of the CLI output-routing hardening work
- Improves harnessability for tests, wrappers, and scripted integrations
- No behavioral change to payload shape beyond writer routing
