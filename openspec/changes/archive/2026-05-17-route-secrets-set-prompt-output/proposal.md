## Why

`lango security secrets set` sends its interactive secret-value prompt through the global passphrase writer instead of the Cobra command writer. Wrappers, tests, and embedded CLI callers that replace `cmd.OutOrStdout()` can capture the success line but miss the prompt, making credential setup look silent or hung.

## What Changes

- Add a stream-aware passphrase helper for hidden input prompts.
- Route the `security secrets set` interactive prompt through `cmd.OutOrStdout()`.
- Add regression tests proving the prompt is visible through the command writer while keeping non-interactive `--value-hex` behavior unchanged.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-secrets-management`: clarify that interactive `secrets set` prompts are part of command output stream routing.

## Impact

- Affected code: `internal/cli/prompt` and `internal/cli/security`.
- No CLI syntax changes.
- Public docs already describe interactive prompting and do not require content changes.
