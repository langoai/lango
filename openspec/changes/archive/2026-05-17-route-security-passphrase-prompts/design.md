## Context

`prompt.PassphraseIO` now lets callers route visible hidden-input prompt text through a supplied writer while still reading the secret from the terminal file descriptor. Several security commands still call `prompt.Passphrase` or `prompt.PassphraseConfirm`, which use package-level stdout seams. Those calls are functionally correct for direct terminal usage but incomplete for Cobra command stream capture.

## Goals / Non-Goals

**Goals:**

- Route visible passphrase prompt text in security commands through Cobra output streams.
- Keep hidden input on the existing terminal password reader path.
- Keep warning and notice output that is already specified for stderr on command error streams.
- Add focused regression coverage without changing crypto or storage behavior.

**Non-Goals:**

- Do not add non-interactive alternatives for these passphrase-changing commands.
- Do not change passphrase acquisition priority during bootstrap.
- Do not rework keyring, recovery, or envelope storage semantics.

## Decisions

- Add `prompt.PassphraseConfirmIO(out, prompt, confirmPrompt)` instead of changing `PassphraseConfirm` signatures. This keeps existing callers compatible and gives command-aware callers an explicit stream contract.
- Update command execution paths to pass `cmd.OutOrStdout()` into hidden prompt helpers. This follows existing `ConfirmDenyOnEOFIO` and `ReadLineIO` patterns already used by the CLI.
- Keep terminal reads global through `term.ReadPassword(passphraseInputFD())`. Visible output routing and hidden input acquisition have different requirements; changing input acquisition would expand scope and risk.
- For `change-passphrase`, pass the output writer into the execution seam rather than reaching for Cobra streams inside tests. This keeps tests deterministic and avoids process stdout capture.

## Risks / Trade-offs

- Package-level test seams remain mutable. Mitigation: keep tests non-parallel and restore seams with `t.Cleanup`.
- Some commands require deeper bootstrapping to reach prompts. Mitigation: test the shared prompt helpers and the command execution seams where possible, and use focused command tests where the command already exposes a seam.
- This improves command stream capture but does not make hidden input non-interactive. That is intentional because these commands protect master credentials and already require an interactive terminal.
