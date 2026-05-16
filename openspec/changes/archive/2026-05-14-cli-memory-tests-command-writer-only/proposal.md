## Why

The memory CLI tests still used the global stdout capture helper for several error-path assertions even though the package already had local command writer helpers.

## What Changes

- Replace `testutil.ExecCmd` usage in memory CLI tests with the package-local command capture helpers
- Keep the same error-path assertions while removing dependence on global stdout interception

## Impact

- Improves test determinism for memory CLI coverage
- Continues the cleanup of legacy stdout-capture helpers from CLI tests
