## Why

The graph CLI tests still used the global stdout capture helper for several success and error-path assertions even though the package already provided a local command-writer capture helper.

## What Changes

- Replace `testutil.ExecCmd`/`ExecCmdOK` usage in graph CLI tests with the package-local command-writer helper
- Keep the same graph CLI assertions while removing reliance on global stdout interception

## Impact

- Improves test determinism for graph CLI coverage
- Reduces another pocket of global stream mutation in CLI tests
