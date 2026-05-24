## Why

The P2P git guidance tests still used the global stdout capture helper even though the package already exposes a local command-writer capture helper.

## What Changes

- Replace `testutil.ExecCmdOK` usage in P2P git tests with the package-local command-writer helper
- Keep the same guidance assertions while removing reliance on global stdout interception

## Impact

- Improves test determinism for P2P git CLI coverage
- Reduces another pocket of global stream mutation in CLI tests
