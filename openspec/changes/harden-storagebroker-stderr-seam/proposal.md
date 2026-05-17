## Why

The storage broker client launches an internal child process with `cmd.Stderr`
bound directly to `os.Stderr`. That bypasses the stream seams used by CLI/TUI
entrypoints and makes broker diagnostics harder to capture deterministically in
tests.

## What Changes

- Route storage broker child-process stderr through an explicit package seam.
- Extract broker command construction into a small helper that can be tested
  without starting a child process.
- Add regression coverage proving the helper uses the injected stderr writer.
- No user-facing CLI behavior or protocol changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `brokered-storage`: Broker child-process stderr routing gains a seam-aware
  guarantee.
- `test-coverage`: Standard test execution gains broker stderr routing
  regression coverage.

## Impact

- Affected code: `internal/storagebroker/client.go`.
- Affected tests: `internal/storagebroker/*_test.go`.
- Affected specs: `brokered-storage`, `test-coverage`.
- No public API, CLI, config, protocol, or dependency changes.
