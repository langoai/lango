## Why

The public docs describe bare `lango` as if it always launches the mission workbench TUI, but the root command intentionally prints help and exits successfully when stdin is non-interactive. That is a production-readiness mismatch for wrappers, CI, Docker checks, and users running with redirected stdin.

## What Changes

- Document that bare `lango` launches the workbench only in an interactive terminal.
- Document that non-interactive bare `lango` prints Cobra help to command stdout and exits successfully without starting the TUI.
- Keep `lango cockpit` and `lango chat` described separately because those commands return actionable non-interactive errors instead of using the bare-root help fallback.
- Add executable docs parity coverage so public docs do not regress.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `downstream-docs-sync`: Public CLI docs must describe the bare-root non-interactive fallback contract.
- `test-coverage`: Docs-runtime parity guards must cover the bare-root non-interactive contract.

## Impact

- Affected docs: `README.md`, `docs/cli/index.md`, and `docs/cli/core.md`.
- Affected tests: docs quality guard coverage in `internal/testutil`.
- Runtime behavior is unchanged.
