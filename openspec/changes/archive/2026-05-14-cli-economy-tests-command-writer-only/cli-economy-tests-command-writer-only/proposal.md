## Why

The economy CLI tests still depended on the global stdout capture helper even though the package already had a local command-writer helper suitable for both guidance and status-path assertions.

## What Changes

- Replace `testutil.ExecCmd` and `ExecCmdOK` usage in economy CLI tests with the package-local command-writer helper
- Keep the same economy CLI assertions while removing reliance on global stdout interception

## Impact

- Improves test determinism for economy CLI coverage
- Reduces another broad cluster of global stream mutation in CLI tests
