## Why

Gateway shutdown is used both for normal server termination and for rollback/cleanup paths. The current lifecycle spec only covers shutdown after a server has started serving, leaving pre-start, failed-start, and repeated shutdown paths under-specified. Production cleanup code should be able to call shutdown safely without needing to know how far gateway startup progressed.

## What Changes

- Make gateway shutdown safe before the HTTP server has started.
- Keep shutdown safe after a startup/listen failure.
- Make repeated shutdown calls deterministic and non-panicking.
- Add focused regression coverage for these lifecycle edges.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `gateway-server`: Gateway shutdown lifecycle safety covers pre-start, failed-start, running-server, and repeated shutdown paths.

## Impact

- Affected code: `internal/gateway/server.go`.
- Affected tests: `internal/gateway/server_test.go`.
- Affected specs: `openspec/specs/gateway-server/spec.md`.
