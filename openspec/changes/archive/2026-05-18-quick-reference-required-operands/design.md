## Scope

This is a docs-and-guard correction. It does not change command behavior.

## Current State

- `lango memory clear` requires `<session-key>` in the actual Cobra command and dedicated memory docs, but README and the CLI index omit it.
- `lango p2p firewall add`, `lango p2p firewall remove`, and `lango p2p session revoke` require peer identifiers in code and detailed P2P docs, but README and the CLI index show bare commands.
- `lango config get` supports `--output plain|json` and `--show-secrets`; the dedicated config docs show the full usage, while README and CLI index only show `--show-secrets`.

## Design

Update the public quick references and affected feature command examples to include required operands and relevant output flags:

- `lango memory clear <session-key>`
- `lango p2p firewall add --peer-did <did>`
- `lango p2p firewall remove <peer-did>`
- `lango p2p session revoke --peer-did <did>`
- `lango config get <dot.path> [--output plain|json] [--show-secrets]`

Strengthen the existing guard tests rather than creating a separate docs-only test harness. The guards should check both README and `docs/cli/index.md` where both files carry the same quick-reference responsibility, check affected feature docs for the specific stale examples in scope, and reject known stale bare rows that previously passed substring checks.

## Verification

- Run the focused guard tests in `internal/testutil`.
- Run `go build ./...`, `go test ./...`, `openspec validate --all --strict`, and `git diff --check` before committing.
