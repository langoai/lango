# Route Interactive Guards Through Command Input

## Why

Several interactive CLI entrypoints already route prompt text through Cobra command
output streams, but their terminal guards still inspect process-global stdin. That
makes command tests and wrappers less reliable because `cmd.SetIn(...)` can drive
later prompts while the initial guard ignores the same input stream.

## What Changes

- Add a shared prompt helper that validates an explicit command input stream for
  interactive-only commands.
- Route onboard, settings, security passphrase, recovery, keyring-store, and
  secrets-set interactive guards through `cmd.InOrStdin()`.
- Preserve existing non-interactive guidance messages and existing prompt output
  routing.

## Impact

- Modified capabilities: `cli-prompt-helpers`, `cli-onboard`, `cli-settings`,
  `cli-secrets-management`, `passphrase-management`, `recovery-mnemonic`.
- Public behavior is unchanged for normal terminal usage.
- Command tests and embedded wrappers can exercise interactive flows through
  explicit streams without process-global stdin interception.
