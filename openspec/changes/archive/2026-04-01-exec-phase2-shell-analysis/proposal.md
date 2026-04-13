## Why

The current shell wrapper unwrap in `unwrapShellWrapper()` uses string-based parsing that cannot handle common shell patterns: login shell flags (`sh -lc`), env wrappers (`/usr/bin/env sh -c`), nested wrappers (`sh -c "bash -c \"inner\""`), and complex quoting. These gaps allow policy bypass through trivial command reformulation.

## What Changes

- Replace string-based unwrap internals with AST-based parser using `mvdan.cc/sh/v3/syntax`
- Support `sh -lc "cmd"` (login shell with `-c`)
- Support `/usr/bin/env sh -c "cmd"` (env wrapper)
- Support recursive unwrap for nested wrappers with depth limit 5
- Fallback to existing string-based parser on AST parse failure
- Keep existing `unwrapShellWrapper(cmd)` function signature unchanged

## Capabilities

### New Capabilities

### Modified Capabilities
- `exec-policy-evaluator`: Add scenarios for login shell, env wrapper, nested wrapper unwrap, and recursive unwrap depth limit

## Impact

- `internal/tools/exec/unwrap.go` — rewrite internals, add AST parser
- `internal/tools/exec/unwrap_test.go` — expanded test coverage for new patterns
- `internal/tools/exec/policy.go` — no signature changes, uses same `unwrapShellWrapper` call
- `internal/tools/exec/policy_test.go` — new test cases for newly-supported wrapper forms
- `prompts/SAFETY.md` — document newly supported wrapper forms
- New dependency: `mvdan.cc/sh/v3` (already added)
