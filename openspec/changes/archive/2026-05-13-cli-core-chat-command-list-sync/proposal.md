## Why

The current cockpit feature docs and in-product `/help` already expose `/mode` and `/cost`, but the `lango cockpit` CLI overview still lists only `/help`, `/clear`, `/model`, `/status`, and `/exit`. That leaves the CLI overview behind the shipped slash-command surface.

## What Changes

- Update `docs/cli/core.md` to include `/mode` and `/cost` in the chat/cockpit slash-command summary.
- Update the downstream-docs-sync spec to require the same CLI-overview command coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: CLI overview slash-command summaries need to stay aligned with the current chat command surface.

## Impact

- Affected docs/specs: `docs/cli/core.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
