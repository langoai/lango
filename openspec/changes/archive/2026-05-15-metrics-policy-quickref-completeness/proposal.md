## Why

The `lango metrics policy` command is implemented and documented in the dedicated metrics reference, but it was still missing from the public quick references. That left one metrics subcommand less discoverable than the rest of the shipped metrics surface.

## What Changes

- add `lango metrics policy` to `README.md`
- add `lango metrics policy` to `docs/cli/index.md`
- widen the existing metrics completeness guard so it enforces the full public quick-reference surface, including `policy`

## Impact

- public quick references better match the actual shipped metrics CLI surface
- reduced operator confusion when skimming metrics commands
- stronger regression protection for metrics quick-reference drift
