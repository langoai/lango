## Why

The configuration system already exposes `config.Provenance`, but the interactive `lango settings` editor does not let users view or edit it. That makes provenance the odd one out among automation-facing subsystems such as Cron, Background, Workflow, and RunLedger.

## What Changes

- Add a `Provenance` category to the Automation section of `lango settings`
- Provide a dedicated provenance settings form for the existing config-backed fields
- Wire form updates into the existing TUI config state flow
- Document that `session_isolation` remains AGENT.md metadata, not a settings field

## Capabilities

### Modified Capabilities

- `cli-settings`: add provenance configuration editing support
- `session-provenance`: expose config-backed editing surface through settings

## Impact

- `internal/cli/settings/`: new provenance form and menu/setup-flow wiring
- `internal/cli/tuicore/`: provenance form field → config mapping
- `openspec/specs/cli-settings/spec.md`: provenance form requirement
- README and provenance docs updated to point users to settings for config-backed provenance behavior
