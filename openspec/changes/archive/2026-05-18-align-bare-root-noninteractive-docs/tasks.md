## 1. Tests

- [x] 1.1 Add a failing docs guard for README, CLI index, and core CLI docs bare-root non-interactive fallback coverage.

## 2. Implementation

- [x] 2.1 Update README command contract and quick reference copy for interactive-only bare `lango`.
- [x] 2.2 Update `docs/cli/index.md` quick reference and narrative with the non-interactive fallback.
- [x] 2.3 Update `docs/cli/core.md` bare `lango` section with the same runtime contract and distinction from `chat`/`cockpit`.

## 3. Verification

- [x] 3.1 Run focused docs guard and existing root command tests.
- [x] 3.2 Run full `go build ./...`, `go test ./...`, `git diff --check`, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Sync specs, archive the change, and commit the scoped unit.
