## Why

The README quick reference still summarized `lango p2p git push` and `lango p2p git fetch` as live bundle actions, but the current CLI surfaces are still server-backed guidance paths. That wording overstates what those commands do today.

## What Changes

- sync the README P2P git quick-reference lines to the current guidance-oriented contract
- add an executable guard so the stale README wording cannot silently return

## Impact

- public quick-reference docs better match the actual CLI UX
- reduced confusion for operators skimming the README command list
- stronger regression protection for README drift
