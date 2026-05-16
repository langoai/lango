## Why

The approval CLI tests still used the global-stdout capture helper even though the command output had already been routed through Cobra writers. That kept the tests more fragile than necessary.

## What Changes

- Replace `testutil.ExecCmd` usage in approval CLI tests with local command-writer capture
- Keep the same approval CLI assertions while removing the need for global stdout interception

## Impact

- Improves test determinism for approval CLI coverage
- Reduces reliance on process-global stream mutation in one more package
