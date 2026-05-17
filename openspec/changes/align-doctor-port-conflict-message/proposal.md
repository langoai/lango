## Why

`lango doctor` currently reports occupied gateway ports as "not available" while the `cli-doctor` spec requires an "in use" diagnostic. The less specific message weakens the out-of-the-box troubleshooting path when a gateway port conflict blocks startup.

## What Changes

- Align the doctor server port conflict message with the existing `cli-doctor` requirement.
- Add a regression test that binds a local port and verifies the occupied-port diagnostic.
- Keep behavior otherwise unchanged; no command flags, output schema, or docs changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-doctor`: Server port checks must report occupied configured ports as "Port <port> in use".
- `test-coverage`: Executable tests must cover the occupied-port doctor diagnostic.

## Impact

- Affected code: `internal/cli/doctor/checks/network.go`.
- Affected tests: `internal/cli/doctor/checks/checks_test.go`.
- Affected specs: `openspec/specs/cli-doctor/spec.md`, `openspec/specs/test-coverage/spec.md`.
