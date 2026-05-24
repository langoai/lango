## Why

The cron CLI tests still used the global stdout capture helper for several error-path assertions even though the command output already follows the Cobra writer contract.

## What Changes

- Replace `testutil.ExecCmd` usage in cron CLI tests with the package-local command capture helper
- Keep the same cron CLI assertions while removing reliance on global stdout interception for these tests

## Impact

- Improves test determinism for cron CLI coverage
- Reduces reliance on process-global stream mutation in another CLI package
