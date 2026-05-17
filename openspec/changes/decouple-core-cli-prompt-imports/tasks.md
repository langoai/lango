# Tasks

## 1. Planning

- [x] 1.1 Audit current non-CLI imports of `internal/cli/prompt`.
- [x] 1.2 Add focused OpenSpec artifacts for core/CLI prompt import decoupling.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [ ] 1.4 Commit the OpenSpec planning artifacts separately.

## 2. Regression Test

- [ ] 2.1 Add an archtest rule that forbids non-CLI `internal/**` production packages from importing `internal/cli/**`.
- [ ] 2.2 Confirm the new archtest fails against the current `internal/security/passphrase` and `internal/bootstrap` imports.

## 3. Implementation

- [ ] 3.1 Remove the `internal/cli/prompt` import from `internal/security/passphrase` while preserving passphrase prompt behavior.
- [ ] 3.2 Remove the `internal/cli/prompt` import from `internal/bootstrap` while preserving confirmation behavior.
- [ ] 3.3 Keep CLI prompt package behavior unchanged.

## 4. Review

- [ ] 4.1 Request teammate review for spec compliance and code quality.
- [ ] 4.2 Address any actionable findings before archiving.

## 5. Verification

- [ ] 5.1 Run focused `internal/archtest`, `internal/security/passphrase`, and `internal/bootstrap` tests.
- [ ] 5.2 Run `go build ./...`.
- [ ] 5.3 Run `go test ./...`.
- [ ] 5.4 Run `openspec validate --all --strict`.
- [ ] 5.5 Run `git diff --check`.
- [ ] 5.6 Archive the change after verification.
