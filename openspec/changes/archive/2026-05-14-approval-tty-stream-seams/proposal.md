## Why

The TTY approval fallback still relied on process-global terminal state, stdin, and stderr. That made the approval prompt path harder to verify than the other recently hardened seam-aware flows.

## What Changes

- Add narrow seams for terminal detection, input, and stderr output in `TTYProvider`
- Add deterministic approval-path tests for approve once, always allow, and deny
- Update docs and OpenSpec with the seam-aware prompt contract

## Impact

- Improves confidence in a security-critical approval fallback path
- Reduces reliance on global process stream mutation in approval tests
