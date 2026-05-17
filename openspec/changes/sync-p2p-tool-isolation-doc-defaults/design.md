# Design: Sync P2P Tool Isolation Doc Defaults

## Decision

Add a focused guard in `internal/testutil/config_docs_quality_guard_test.go` that reads `README.md` and `docs/configuration.md`, parses markdown table rows with the existing helpers, and asserts selected P2P tool isolation defaults match `config.DefaultConfig()`.

## Rationale

`DefaultConfig()` is already specified as the single source of truth for configuration defaults. A guard tied directly to `DefaultConfig()` prevents public docs from drifting when defaults change, without requiring a broad documentation generator or a larger docs refactor.

## Verification Strategy

- First add the guard and confirm it fails against the stale README row.
- Update README only after the failing guard proves the drift.
- Run focused `internal/testutil` tests, full build/test, OpenSpec validation, and whitespace checks.

## Compatibility

This change only affects public documentation and tests. It does not alter runtime defaults, CLI behavior, config loading, or sandbox execution.
