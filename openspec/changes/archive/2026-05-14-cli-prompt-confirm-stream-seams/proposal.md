## Why

`prompt.Confirm(...)` still hard-codes `os.Stdin` and `os.Stdout`, even though adjacent prompt helpers already expose deterministic seams for tests. That leaves the package-default confirmation path harder to verify and inconsistent with the repository's broader command-stream hardening work.

## What Changes

- Add small injectable seams for the default confirmation input and output streams used by `prompt.Confirm(...)`
- Add prompt package regressions that exercise the wrapper's approve, deny, and read-error paths without replacing process-global stdio
- Restore the shared config/bootstrap loader helpers required by existing CLI regression tests after the global stdout harness removal
- Record the confirmation-wrapper seam contract in OpenSpec

## Capabilities

### New Capabilities
- `cli-prompt-helpers`: Shared prompt helpers expose deterministic stream seams for default confirmation flows

### Modified Capabilities
- `test-coverage`: Shared CLI test helpers remain available after removing the global stdout interception harness

## Impact

- Affected code: `internal/cli/prompt/prompt.go`, `internal/cli/prompt/prompt_test.go`, `internal/testutil/*`
- Affected specs: new `cli-prompt-helpers` capability and updated `test-coverage`
- No runtime CLI behavior changes; this is a testability and maintenance improvement
