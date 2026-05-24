## 1. Vault Input Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing required crypto inputs.
- [x] 1.2 Add tool-entrypoint regression coverage for missing required secrets inputs.

## 2. Downstream Sync

- [x] 2.1 Update tool-usage prompt wording for the vault security input contract.
- [x] 2.2 Update README and CLI docs for the same contract.
- [x] 2.3 Update security and production-readiness specs for the required-input coverage.

## 3. Verification

- [x] 3.1 Run `go test ./internal/tools/crypto ./internal/tools/secrets -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
