## Context

`prompt.Passphrase` writes prompt text and the trailing newline to a package-level `passphraseOutput` defaulting to `os.Stdout`. This is acceptable for global passphrase prompts, but `security secrets set` is a Cobra command whose spec requires output to be routed through the command writer.

`secrets delete` already uses command input/output streams for confirmation. `secrets set` should follow the same output routing for its visible prompt while retaining hidden input via the terminal password reader.

## Decision

Introduce `prompt.PassphraseIO(out io.Writer, prompt string)`:

- It writes the prompt and trailing newline to the provided writer.
- It reuses the existing hidden input reader and input file descriptor.
- `prompt.Passphrase` delegates to `PassphraseIO(passphraseOutput, prompt)` to preserve existing behavior.

`newSecretsSetCmd` will call `prompt.PassphraseIO(cmd.OutOrStdout(), "Enter secret value: ")` for the interactive branch.

## Test Strategy

- Add prompt package unit tests for `PassphraseIO` output routing and newline behavior.
- Add a secrets command test using package seams to avoid a real terminal or real password reader, proving the prompt appears in the Cobra output buffer.
- Keep existing non-interactive hex tests as coverage that `--value-hex` bypasses the prompt.
