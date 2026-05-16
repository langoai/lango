## Why

The CLI quick reference recently gained several hidden-but-implemented operator commands, and the page also had a malformed Agent & Memory table because prose was embedded between rows. Both issues should be guarded mechanically.

## What Changes

- fix the Agent & Memory table formatting on `docs/cli/index.md`
- add an executable docs guard for implemented KMS wrap/detach and P2P workspace/provenance quick-reference entries
- add an executable guard preventing prose from splitting the Agent & Memory table
- sync docs-only and test-coverage specs

## Impact

- keeps the public quick reference structurally valid
- preserves discoverability for implemented operator commands
