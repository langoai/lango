# Design: Decouple Core CLI Prompt Imports

## Decision

Use the existing import-graph archtest to enforce the boundary, then remove the current violations by replacing direct `internal/cli/prompt` imports in security/bootstrap with package-local prompt seams built on lower-level primitives.

## Boundary

The new rule applies to production packages under `internal/**` except `internal/cli/**`. Those packages SHALL NOT import `github.com/langoai/lango/internal/cli` or any subpackage. `cmd/lango` remains allowed to import CLI packages because it is the command entrypoint.

## Implementation Notes

- `internal/security/passphrase` already has prompt seams; their defaults can be implemented locally with `term.ReadPassword` and explicit writer routing.
- `internal/bootstrap` only needs a yes/no confirmation seam for storing passphrases; it can use `lineio.ReadLine` directly without depending on CLI prompt helpers.
- Tests should first demonstrate the current boundary violation in `go test ./internal/archtest`.

## Compatibility

Runtime behavior should remain unchanged: passphrase acquisition still prefers keyring, keyfile, interactive terminal, then stdin; bootstrap still asks for confirmation before storing interactive passphrases.
