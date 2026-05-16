## Why

The P2P team CLI tests still used the global stdout capture helper for guidance and help assertions even though the package already exposes a local command-writer helper.

## What Changes

- Replace `testutil.ExecCmdOK` usage in P2P team tests with the package-local command-writer helper
- Keep the same guidance/help assertions while removing reliance on global stdout interception

## Impact

- Improves test determinism for P2P team CLI coverage
- Reduces another pocket of global stream mutation in CLI tests
