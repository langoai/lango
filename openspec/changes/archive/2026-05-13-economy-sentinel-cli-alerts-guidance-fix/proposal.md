## Why

`lango economy escrow sentinel status` still tells operators to run `lango economy escrow sentinel alerts`, but no such CLI subcommand exists. The real follow-up surface is the running server plus the `sentinel_alerts` agent tool. The CLI and docs should stop pointing users to a nonexistent command.

## What Changes

- Replace the stale `lango economy escrow sentinel alerts` guidance with truthful `sentinel_alerts` agent-tool guidance.
- Add a CLI regression that locks the new message.
- Sync the economy CLI docs and sentinel spec to the same contract.

## Impact

- `escrow-sentinel`: CLI status guidance becomes truthful and actionable.
- Operator UX: users are no longer sent to a nonexistent sentinel alerts subcommand.
