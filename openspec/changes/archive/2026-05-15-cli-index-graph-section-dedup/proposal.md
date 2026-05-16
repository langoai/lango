## Why

The repository now ships a dedicated graph CLI reference, but the public CLI index still embeds graph quick-reference rows inside the `Agent & Memory` section. That keeps the index structure out of sync with the dedicated docs split and makes graph discoverability look like an agent-memory subtopic instead of its own command family.

## What Changes

- move `lango graph ...` rows into a dedicated `Graph Store` section in `docs/cli/index.md`
- leave an explicit handoff from the `Agent & Memory` section to `docs/cli/graph.md`
- add an executable guard so graph quick-reference rows do not drift back into the wrong section

## Impact

- cleaner CLI index structure that matches the dedicated docs inventory
- less scope confusion between graph commands and memory commands
- stronger regression protection for future CLI docs edits
