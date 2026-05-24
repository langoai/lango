## Why

Gateway-backed CLI commands should report and call the exact normalized gateway target chosen by the operator. After config-backed address resolution, `lango status --addr <url>` can probe the override while still displaying the configured gateway, and explicit addresses with a trailing slash can produce double-slash request paths.

## What Changes

- normalize explicit gateway addresses by trimming trailing slashes in the shared CLI resolver
- make `lango status --addr <url>` display the same normalized gateway that it probes
- add focused regression tests for explicit status target display and trailing-slash gateway normalization
- keep configured gateway defaults unchanged for metrics, alerts, bg, and status

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `cli-status-dashboard`: status output reports the actual normalized gateway target used for live probing
- `test-coverage`: executable tests guard explicit target display and trailing-slash normalization

## Impact

- affects shared gateway address resolution in `internal/cli/clihttp`
- affects status collection/display in `internal/cli/status`
- may slightly change JSON/table output for `lango status --addr <url>` from configured address to explicit target, which is the more truthful operator-facing behavior
