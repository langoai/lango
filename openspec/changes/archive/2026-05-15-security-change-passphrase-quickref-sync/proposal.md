## Why

The current security command surface includes canonical `lango security change-passphrase` alongside the deprecated `migrate-passphrase` path, but the public quick references still surfaced only the legacy command. That made the operator-facing docs lag behind the actual migration path.

## What Changes

- add `lango security change-passphrase` to `README.md`
- add `lango security change-passphrase` to `docs/cli/index.md`
- widen the existing security completeness guard so it enforces the current public quick-reference surface across both files

## Impact

- public quick references better match the actual shipped security CLI surface
- operators see the canonical passphrase-rotation path immediately
- stronger regression protection for security quick-reference drift
