# Proposal: Harden Passphrase Public Stdio Seams

## Summary

Route the public passphrase acquisition wrappers through package-level stdio seams so tests can exercise wrapper behavior without replacing process-global standard streams.

## Problem

`Acquire` and `AcquireNonInteractive` currently bind directly to `os.Stdin`, `os.Stderr`, and terminal detection. Lower-level helpers already accept injected readers and writers, but the public wrappers still make focused wrapper regression tests depend on process-global stdio.

## Goals

- Keep passphrase acquisition behavior unchanged.
- Make `Acquire` testable with injected stdin, stderr, and terminal detection seams.
- Make `AcquireNonInteractive` warning output testable through the same stderr seam.
- Add focused tests that prove the public wrappers use the seams.

## Non-Goals

- Changing passphrase source priority.
- Changing interactive prompt text.
- Changing keyring, keyfile, or stdin parsing behavior.
