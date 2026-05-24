## Why

`lango doctor` now reports every server listen failure as a port conflict. That is correct when another process already owns the configured port, but it is misleading when `server.host` is invalid or not assignable on the machine. The diagnostic should tell users whether to free a port or fix their bind address.

## What Changes

- Keep occupied-port failures reported as `Port <port> in use`.
- Report non-conflict listen failures as a server bind-address failure while preserving the original system error in details.
- Add a regression test for a malformed bind host so future doctor changes do not mislabel configuration errors as port conflicts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-doctor`: Server port checks distinguish occupied ports from invalid or unavailable bind addresses.
- `test-coverage`: Executable tests cover non-conflict doctor listen failures.

## Impact

- Affected code: `internal/cli/doctor/checks/network.go`.
- Affected tests: `internal/cli/doctor/checks/checks_test.go`.
- Affected specs: `openspec/specs/cli-doctor/spec.md`, `openspec/specs/test-coverage/spec.md`.
