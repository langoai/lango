## Why

The economy CLI tests still depended on the global stdout capture helper even though the package already exposes a local command-writer helper for ordinary command execution.

## What Changes

- Replace `testutil.ExecCmd`/`ExecCmdOK` usage in economy CLI tests with the package-local command-writer helper
- Keep the same help, guidance, and status assertions while removing reliance on global stdout interception

## Impact

- Improves test determinism for economy CLI coverage
- Reduces another cluster of global stream mutation in CLI tests
