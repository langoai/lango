## Context

Dead Letters already fail-closes `Ctrl+R` while `retryRunning()` is true, but `ShortHelp()` still renders the binding unconditionally in the base help set. This is a help-surface problem rather than a retry execution problem.

## Goals / Non-Goals

**Goals:**

- Remove inert reset help while retry work is active.
- Preserve the existing reset binding in all other Dead Letters states.

**Non-Goals:**

- Change retry execution semantics.
- Change filter reset behavior outside retry-running state.

## Decisions

- Gate the base `ctrl+r` help binding on `!retryRunning()`.
  - Rationale: this mirrors the existing runtime key guard exactly.

## Risks / Trade-offs

- [Help count changes during active retry] → This is intentional because the goal is to show only actionable controls.
