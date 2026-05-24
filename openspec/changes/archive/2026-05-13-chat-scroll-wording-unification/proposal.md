## Why

The chat surface currently describes `PgUp/PgDn` with two different phrases: `/help` says "Scroll chat" while the cockpit feature docs say "Scroll the transcript viewport". The behavior is the same, but the wording split makes the surface feel less intentional than it should.

## What Changes

- Unify `PgUp/PgDn` wording between `/help` and the public cockpit feature docs.
- Add a small regression for the `/help` copy.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Chat help text for transcript scrolling uses a single wording contract.

## Impact

- Affected code: `internal/cli/chat/commands.go`, `internal/cli/chat/chat_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/tui-chat-rendering/spec.md`
