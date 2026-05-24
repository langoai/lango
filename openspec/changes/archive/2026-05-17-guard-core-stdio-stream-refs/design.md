# Design

## Approach

Add a test file under `internal/testutil` that walks `internal/` production Go files and rejects direct `os.Stdin`, `os.Stdout`, or `os.Stderr` references unless the file/line is explicitly allowlisted.

The guard will skip:

- `internal/cli/`, because `cli_stream_quality_guard_test.go` already owns that layer.
- `internal/testutil/`, because these tests contain fixtures and guard strings.
- generated `internal/ent/` code and comments.

The allowlist will be line-fragment based, matching existing cmd entrypoint guard style. Allowed entries should represent intentional seams or generated/template comments only.

## Initial Allowlist Categories

- Existing package-level stdio seams: approval TTY, passphrase acquisition, sandbox worker, logging, broker stderr, tracing, bootstrap, exec warnings.
- Passphrase stdin helper public wrapper while lower-level `ReadStdinPipeFromReader` remains the injectable path.
- Generated ent migration comment that references `os.Stdout`.

## Test Strategy

Use TDD:

1. Add a fixture test that calls a not-yet-implemented scanner and expects a direct `os.Stdout` reference to be rejected.
2. Implement the scanner and real repository guard.
3. Run the focused guard package, full build, full test suite, and OpenSpec validation.

## Non-Goals

- Do not change production stdio behavior in this change.
- Do not scan `cmd/` or `internal/cli/`; existing guards own those areas.
- Do not remove intentional seams.
