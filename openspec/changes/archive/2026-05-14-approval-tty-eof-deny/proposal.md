## Why

The TTY approval fallback still surfaces EOF as a read error even though the rest of the interactive confirmation stack now treats missing confirmation input as a safe denial. That makes a sensitive approval path behave more harshly and less consistently than the CLI confirmation flows around it.

## What Changes

- Treat EOF on the TTY approval fallback as a clean denial instead of a read error
- Add a regression covering seam-injected EOF on the TTY provider
- Document the EOF-deny contract in the approval docs and approval spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `channel-approval`: TTY fallback treats EOF as a safe denial

## Impact

- Affected code: `internal/approval/tty.go`, `internal/approval/tty_test.go`
- Affected docs: `docs/security/approval-cli.md`
- Affected specs: `openspec/specs/channel-approval/spec.md`
- No feature expansion; this is approval hardening and behavior consistency
