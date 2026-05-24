## Why

The `lango p2p team` and `lango p2p workspace` command groups already use more truthful runtime guidance in their command outputs, but their top-level help text still uses vague phrases like "agent/tool-backed control paths" or "agent/tool-backed flows". For operators reading `--help`, those descriptions are less actionable than the concrete `team_*` and `p2p_workspace_*` tool surfaces that now exist.

## What Changes

- Replace vague top-level `p2p team` / `p2p workspace` help wording with concrete tool-surface references.
- Add CLI regressions covering the updated help output.
- Sync the public P2P CLI docs and CLI P2P management spec to the same concrete help contract.

## Impact

- `cli-p2p-management`: help output becomes more actionable and less hand-wavy.
- Operator UX: `--help` points directly to the concrete tool surfaces that exist today.
