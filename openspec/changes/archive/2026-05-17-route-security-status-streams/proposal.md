# Route Security Status Streams

## Why

`lango security status` already routes rendered table/JSON output through Cobra
command stdout, but the non-interactive DB-status helper writes passphrase
warnings through a package-global stderr seam. That means command wrappers cannot
capture all status diagnostics through `cmd.ErrOrStderr()`, and tests still need
process-level seams for warning capture.

The same helper also calls secure-provider detection directly, which makes the
keyring-provider path harder to verify without touching platform-specific
backends.

## What Changes

- Route non-interactive status warnings through the command error stream.
- Thread an explicit warning writer into the status DB helper.
- Add a secure-provider detection seam for status non-interactive acquisition.
- Add tests proving command stderr capture and provider seam wiring.

## Impact

- Modified capability: `cli-security-status`.
- No user-facing status text changes.
- No changes to hidden passphrase input or bootstrap behavior.
