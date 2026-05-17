## 1. Tests First

- [x] 1.1 Add a failing prompt test proving a hidden-input prompt can be routed to an explicit writer.
- [x] 1.2 Add a failing `security secrets set` command test proving the interactive prompt is captured through Cobra output.
- [x] 1.3 Run focused tests and confirm they fail before implementation.

## 2. Implementation

- [x] 2.1 Add `prompt.PassphraseIO(out, prompt)` and make `Passphrase` delegate to it.
- [x] 2.2 Route `security secrets set` interactive prompt through `cmd.OutOrStdout()`.
- [x] 2.3 Keep `--value-hex` non-interactive behavior unchanged.

## 3. Verification

- [x] 3.1 Run focused prompt and security CLI tests.
- [x] 3.2 Run `openspec validate route-secrets-set-prompt-output --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent review.
- [x] 3.5 Sync and archive the OpenSpec change.
