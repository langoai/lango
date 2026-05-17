# Proposal: Decouple Core CLI Prompt Imports

## Why

Security and bootstrap packages currently import `internal/cli/prompt` for terminal passphrase and confirmation prompts. That creates an upward dependency from non-CLI production code into CLI UI helpers, which makes headless, daemon, and test runtimes inherit CLI prompt plumbing indirectly.

## What Changes

- Add an archtest guard that forbids non-CLI `internal/**` production packages from importing `internal/cli/**`.
- Remove the existing `internal/cli/prompt` imports from `internal/security/passphrase` and `internal/bootstrap`.
- Preserve existing interactive passphrase and bootstrap confirmation behavior through local seams or lower-level non-CLI utilities.
- Keep CLI prompt package behavior unchanged for CLI callers.

## Non-Goals

- Redesigning the full bootstrap credential flow.
- Changing passphrase source priority or persisted credential behavior.
- Removing interactive terminal support.
