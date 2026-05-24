# Tasks

## 1. Planning

- [x] 1.1 Audit current non-CLI imports of `internal/cli/prompt`.
- [x] 1.2 Add focused OpenSpec artifacts for core/CLI prompt import decoupling.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [x] 1.4 Commit the OpenSpec planning artifacts separately.

## 2. Regression Test

- [x] 2.1 Add an archtest rule that forbids non-CLI `internal/**` production packages from importing `internal/cli/**`.
- [x] 2.2 Confirm the new archtest fails against the current `internal/security/passphrase` and `internal/bootstrap` imports.

## 3. Implementation

- [x] 3.1 Remove the `internal/cli/prompt` import from `internal/security/passphrase` while preserving passphrase prompt behavior.
- [x] 3.2 Remove the `internal/cli/prompt` import from `internal/bootstrap` while preserving confirmation behavior.
- [x] 3.3 Keep CLI prompt package behavior unchanged.

## 4. Review

- [x] 4.1 Request teammate review for spec compliance and code quality.
- [x] 4.2 Address any actionable findings before archiving.

## 5. Verification

- [x] 5.1 Run focused `internal/archtest`, `internal/security/passphrase`, and `internal/bootstrap` tests.
- [x] 5.2 Run `go build ./...`.
- [x] 5.3 Run `go test ./...`.
- [x] 5.4 Run `openspec validate --all --strict`.
- [x] 5.5 Run `git diff --check`.
- [x] 5.6 Archive the change after verification.
