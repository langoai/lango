## Why

The public P2P feature page still summarized `lango p2p git push` and `lango p2p git fetch` as direct live git execution, but the current CLI surfaces are still server-backed guidance paths. That wording overstates what the commands do today.

## What Changes

- sync the P2P feature-page git command summaries to the current server-backed guidance contract
- add an executable docs guard so the stale git summary wording cannot silently return

## Impact

- public docs better match the actual CLI UX
- reduced confusion for operators scanning the P2P command list
- stronger regression protection for feature-doc drift
