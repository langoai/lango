## 1. Tests

- [x] 1.1 Add failing CLI tests for `lango cron add --timeout` and `--deliver-to`.
- [x] 1.2 Add failing CLI tests for cron control `--id` selectors and ambiguous selector rejection.
- [x] 1.3 Add failing CLI test for invalid `--timeout`.
- [x] 1.4 Add failing CLI tests for documented isolated default and idempotent add behavior.

## 2. Implementation

- [x] 2.1 Parse and store `--timeout` in `newAddCmd`.
- [x] 2.2 Add `--deliver-to` as an alias for `--deliver`.
- [x] 2.3 Add `--id` selector support to delete, pause, resume, and history without breaking positional selectors.
- [x] 2.4 Preserve existing output routing and command behavior.
- [x] 2.5 Align CLI add with documented isolated default and store upsert behavior.

## 3. Documentation

- [x] 3.1 Update public cron docs to use accepted flags and selectors.
- [x] 3.2 Sync OpenSpec main specs with the CLI/docs behavior.

## 4. Verification

- [x] 4.1 Validate the OpenSpec change in strict mode.
- [x] 4.2 Run focused cron CLI tests.
- [x] 4.3 Run `go build ./...` and `go test ./...`.
- [x] 4.4 Run subagent-driven review.
- [x] 4.5 Archive the OpenSpec change.
- [x] 4.6 Commit this scoped unit separately.
