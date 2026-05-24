## Why

The `lango p2p team` guidance commands still use vague phrases like "team runtime or agent tools" instead of pointing operators to the actual `team_*` tool surface that exists today. The docs already enumerate the canonical tool-backed path, so the CLI guidance should be just as concrete.

## What Changes

- Replace vague `p2p team` guidance strings with concrete `team_form`, `team_status`, `team_list`, and `team_disband` tool references.
- Add CLI regressions covering the new guidance strings.
- Sync the public P2P CLI docs and CLI P2P management spec to the same concrete operator contract.

## Impact

- `cli-p2p-management`: team guidance becomes specific and easier to act on.
- Operator UX: users are pointed to real tool names instead of generic runtime wording.
