## Why

The doctor network availability regression still uses a hard-coded high port. That can fail intermittently on developer machines or CI when another process owns the same port, even though the doctor implementation is correct.

## What Changes

- Allocate a free local TCP port for the doctor port-available regression.
- Assert the expected success message and listen-address details for the dynamically allocated port.
- Keep production behavior unchanged.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `test-coverage`: Doctor network tests avoid fixed port assumptions for available-port checks.

## Impact

- Affected tests: `internal/cli/doctor/checks/checks_test.go`.
- Affected specs: `openspec/specs/test-coverage/spec.md`.
