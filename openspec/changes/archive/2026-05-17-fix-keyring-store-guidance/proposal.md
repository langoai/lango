# Fix Keyring Store Guidance

## Why

`lango security change-passphrase` and `lango security recovery restore` warn
when updating the secure keyring fails, but the manual recovery guidance points
to `lango security keyring set`. That subcommand does not exist; the supported
command is `lango security keyring store`.

This creates a dead-end precisely when the user is recovering from a stale
credential state that can break headless bootstrap.

## What Changes

- Point stale keyring update warnings to `lango security keyring store`.
- Share the warning text between passphrase change and recovery restore so the
  guidance cannot drift again.
- Add regression coverage for the warning text.

## Impact

- Modified capability: `passphrase-management`.
- User-facing warning text changes only for keyring update failure guidance.
- No cryptographic, storage, or keyring backend behavior changes.
