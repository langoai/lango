## Why

`lango doctor` checks server port availability by manually joining `server.host` and `server.port`. That duplicates gateway listen-address logic and breaks IPv6 hosts such as `::1`, making doctor report false failures for valid gateway configuration.

## What Changes

- Route the doctor server port check through the shared gateway listen-address formatter.
- Add regression coverage for IPv6 and bracketed IPv6 doctor network checks.
- Keep the change focused on address formatting; no CLI surface or configuration schema changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-doctor`: Server port checks must use bracket-safe host/port formatting.
- `gateway-server`: Gateway listen-address formatting must be reused by doctor server port checks.
- `test-coverage`: Executable tests must cover doctor server port checks with IPv6 hosts.

## Impact

- Affected code: `internal/cli/doctor/checks/network.go`.
- Affected tests: `internal/cli/doctor/checks/checks_test.go`.
- Affected specs: `openspec/specs/cli-doctor/spec.md`, `openspec/specs/gateway-server/spec.md`, `openspec/specs/test-coverage/spec.md`.
