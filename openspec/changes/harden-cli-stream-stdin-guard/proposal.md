## Why

The CLI production stream guard rejects direct `os.Stdout` and `os.Stderr` references under `internal/cli`, but the test-coverage spec requires guarding forbidden direct standard-stream references generally. A production CLI file could reintroduce direct `os.Stdin` access without the current guard failing.

## What Changes

- Add a regression test that proves the stream guard rejects direct `os.Stdin` usage in production CLI files.
- Reuse a small scanner helper so the production guard and fixture regression share the same logic.
- Preserve the existing approved exceptions for prompt/security stream seams.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `test-coverage`: CLI production stream guards cover direct `os.Stdin` references in addition to stdout and stderr.

## Impact

- Affected tests: `internal/testutil/cli_stream_quality_guard_test.go`.
- Affected specs: `openspec/specs/test-coverage/spec.md`.
