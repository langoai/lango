## Why

The README internal CLI inventory still compresses the P2P family to broad buckets like `session/sandbox/workspace/git/provenance/team/zkp`, even though the repository already documents the real subcommand slices in the CLI index and P2P feature docs.

## What Changes

- expand the README internal tree P2P row to the current subcommand slices
- sync the existing P2P inventory guard and main specs with that more truthful inventory wording

## Impact

- more truthful README inventory docs
- better operator discoverability for the shipped P2P surface
- stronger regression protection against broad family-only shorthand
