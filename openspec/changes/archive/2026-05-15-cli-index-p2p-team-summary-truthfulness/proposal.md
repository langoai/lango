## Why

The public CLI quick reference still summarized `lango p2p team list/status/disband` as direct live-control commands, but the current team CLI family is a guidance surface for the running server and team tools. That wording overstates the current operator UX.

## What Changes

- sync the CLI index summaries for the P2P team family to the current guidance-oriented contract
- extend the CLI index docs guard so the stale live-control wording cannot silently return

## Impact

- public quick reference better matches the actual CLI behavior
- reduced operator confusion when scanning the P2P command table
- stronger regression protection for CLI index drift
