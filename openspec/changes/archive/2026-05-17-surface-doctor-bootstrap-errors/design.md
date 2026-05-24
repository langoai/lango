# Design

## Root Cause

`internal/cli/doctor.run` captures the bootstrap error only in a local variable and drops it after deciding not to set `cfg`. The downstream `ConfigCheck` can only infer that no config loaded, so it cannot report the original bootstrap failure.

## Approach

- Append a synthetic `checks.Result` before normal checks when bootstrap returns an error.
- Use a stable check name such as `Bootstrap`.
- Set status to `StatusFail`, message to `Bootstrap failed`, and details to the wrapped bootstrap error.
- Keep normal checks running after that result to preserve existing best-effort doctor behavior.
- Guard `boot.Close()` behind a non-nil bootstrap result.

## Non-Goals

- Do not make bootstrap failures abort the doctor command before rendering.
- Do not change how individual checks handle `cfg == nil`.
- Do not change bootstrap internals.
