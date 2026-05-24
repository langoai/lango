## Context

The passphrase acquisition package owns the source priority chain used by bootstrap. It exposes `acquireWithIO` so tests can inject stdin and stderr, but the interactive path currently calls `prompt.Passphrase` and `prompt.PassphraseConfirm`, which use package-global stdout seams. Recent CLI prompt work added `PassphraseIO` and `PassphraseConfirmIO`, so acquisition can keep hidden input unchanged while routing visible prompts through its injected writer.

## Goals / Non-Goals

**Goals:**

- Make the interactive acquisition branch honor the writer already passed to `acquireWithIO`.
- Keep keyring warning output and visible prompt text on the same acquisition diagnostics stream.
- Preserve keyring, keyfile, interactive, and stdin priority order.

**Non-Goals:**

- Do not introduce command-specific bootstrap loaders in this change.
- Do not change terminal file descriptor handling for hidden password reads.
- Do not change `AcquireNonInteractive`.

## Decisions

- Add an internal helper that accepts `out io.Writer` and calls `prompt.PassphraseIO` or `prompt.PassphraseConfirmIO`. This is the smallest change that fixes the seam without changing exported APIs.
- Keep `Acquire` defaulting to `os.Stderr` for acquisition diagnostics and visible prompt text. Passphrase prompts are operational interaction, not command data, so stderr is the safer default than stdout.
- Keep tests at the package seam level. Full bootstrap command stream plumbing is a separate, larger CLI architecture task.

## Risks / Trade-offs

- `Acquire` still cannot see a Cobra command writer directly. Mitigation: this change prevents stdout pollution and makes lower-level behavior testable; command-aware bootstrap wiring can build on this later.
- Tests use prompt package globals for hidden input seams indirectly. Mitigation: tests restore seams with cleanup and do not run in parallel.
