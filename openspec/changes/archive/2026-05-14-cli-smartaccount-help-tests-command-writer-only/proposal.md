## Why

The smartaccount root help regression still used the global stdout capture helper even though the command only needs ordinary Cobra writer capture.

## What Changes

- Replace `testutil.ExecCmd` usage in the smartaccount root help test with a package-local command writer helper

## Impact

- Improves test determinism for smartaccount CLI help coverage
- Removes one more dependency on process-global stdout interception
