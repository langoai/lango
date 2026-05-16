## Why

Some public docs still describe advanced configuration through `lango onboard` or through a nonexistent plaintext `~/.lango/config.yaml` file. Those references are inconsistent with the current product, which stores configuration in the encrypted profile database and exposes advanced editing through `lango settings` or `lango config import/export`.

## What Changes

- Replace remaining advanced-feature README guidance that incorrectly points to `lango onboard`.
- Remove leftover plaintext `config.yaml` wording from provider and alerting docs.
- Align advanced configuration guidance with `lango settings` and `lango config import/export`.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: Keep advanced configuration docs aligned with encrypted-profile storage and the real interactive/programmatic config paths.

## Impact

- Affected docs: `README.md`, `docs/features/ai-providers.md`, `docs/cli/alerts.md`
- No code, API, or runtime behavior changes
