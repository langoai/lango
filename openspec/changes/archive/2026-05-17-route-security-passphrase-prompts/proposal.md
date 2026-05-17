## Why

Several security commands still print hidden-input prompt text through process-global stdout seams. That makes command wrappers, tests, and embedded CLI callers unable to capture the full user interaction through Cobra command streams.

## What Changes

- Add an explicit-output variant for passphrase confirmation prompts.
- Route security passphrase and mnemonic prompt text through `cmd.OutOrStdout()` where a Cobra command is available.
- Preserve hidden terminal input behavior for passphrase reads and preserve warning output on `cmd.ErrOrStderr()`.
- Add regression tests for command-output prompt capture.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-prompt-helpers`: Add explicit-output hidden passphrase helper contracts.
- `passphrase-management`: Require security passphrase-changing prompts to use Cobra command output streams.
- `recovery-mnemonic`: Require recovery setup and restore passphrase prompts to use Cobra command output streams.

## Impact

- Affected code: `internal/cli/prompt`, `internal/cli/security`.
- Affected tests: prompt helper tests and security command tests.
- No CLI flag, config, storage, crypto, or file format changes.
