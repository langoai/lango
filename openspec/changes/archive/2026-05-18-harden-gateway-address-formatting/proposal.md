## Why

Gateway HTTP addresses are assembled in several places with local `fmt.Sprintf("http://%s:%d", host, port)` or `%s:%d` formatting. That duplicates policy and breaks for IPv6 hosts such as `::1`, where the valid URL form is `http://[::1]:18789`. P2P provenance also bypasses the shared CLI resolver, so explicit `--addr http://host/` can produce double-slash gateway paths.

## What Changes

- introduce a small shared gateway address formatting helper
- use bracket-safe host/port formatting for client URLs and server listen addresses
- route P2P provenance push/fetch through the shared CLI gateway resolver
- preserve configured `server.host` and `server.port` fallback when `--addr` is omitted
- keep wildcard bind display behavior stable while using loopback for doctor reachability probes
- document the P2P provenance `--addr` behavior in public CLI docs
- add focused regression tests for IPv6 formatting and P2P provenance normalization/fallback

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `cli-p2p-management`: P2P provenance gateway commands use the shared gateway address contract
- `gateway-server`: gateway listen and display address formatting handles IPv6 host syntax
- `downstream-docs-sync`: P2P provenance docs describe configured fallback and explicit override behavior
- `test-coverage`: executable tests cover P2P provenance gateway address normalization

## Impact

- affects `internal/cli/p2p` provenance push/fetch address handling
- affects shared CLI gateway address formatting
- affects gateway server listen address formatting
- affects serve startup summary and doctor websocket reachability URL formatting
- affects public `docs/cli/p2p.md`
- no API endpoint changes
