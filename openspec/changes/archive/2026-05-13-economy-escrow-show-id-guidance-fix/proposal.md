## Why

`lango economy escrow show --id` currently tells operators to use vague "escrow agent tools" for live data. The actual canonical runtime surface is more specific: on-chain live escrow inspection goes through the running server plus the `escrow_status` agent tool. The CLI and docs should use that concrete path.

## What Changes

- Replace the vague `escrow agent tools` guidance in `lango economy escrow show --id` with truthful `escrow_status` guidance.
- Add a CLI regression covering the new message.
- Sync the economy CLI docs and on-chain escrow spec to the same operator contract.

## Impact

- `onchain-escrow`: operator guidance becomes concrete instead of vague.
- CLI UX: users are sent to the real live-status tool surface.
