## Why

The learning and librarian CLI tests still used the global stdout capture helper for a few error-path assertions even though both command groups already expose package-local command writer helpers.

## What Changes

- Replace `testutil.ExecCmd` usage in learning CLI tests with local command-writer capture
- Replace `testutil.ExecCmd` usage in librarian CLI tests with local command-writer capture

## Impact

- Improves test determinism for both CLI packages
- Reduces reliance on process-global stream mutation in more regression suites
