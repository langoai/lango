## Why

`lango config get`, `set`, and `keys` are implemented and specified, but they are effectively missing from the public CLI docs. That leaves a gap between the real CLI surface and the operator-facing reference.

## What Changes

- add public CLI documentation sections for `config get`, `config set`, and `config keys`
- add the commands to the CLI index config-management table

## Impact

- improves discoverability of the config CLI surface
- aligns public docs with the implemented command set
