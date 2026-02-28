## 1. Wire bg CLI Command

- [x] 1.1 Add `background` and `clibg` imports to `cmd/lango/main.go`
- [x] 1.2 Register `lango bg` command with stub manager provider and GroupID "infra"
- [x] 1.3 Verify `go build ./...` succeeds

## 2. Update README.md

- [x] 2.1 Update Features section security line with keyring, SQLCipher, Cloud KMS
- [x] 2.2 Add security keyring/db-migrate/db-decrypt/kms commands to CLI Commands section
- [x] 2.3 Add p2p session and sandbox commands to CLI Commands section
- [x] 2.4 Add bg commands to CLI Commands section
- [x] 2.5 Add dbmigrate, lifecycle, keyring, sandbox, cli/p2p packages to Architecture tree
- [x] 2.6 Update cli/security description to include new subcommands
- [x] 2.7 Correct skills description (remove "38 embedded", add removal explanation)
- [x] 2.8 Update Skill System description in detailed features section

## 3. Update docs/cli/index.md

- [x] 3.1 Add 8 security extension commands to Security table
- [x] 3.2 Add P2P Network section with 17 commands between Payment and Automation
- [x] 3.3 Add 4 bg commands to Automation section

## 4. Update docs/index.md

- [x] 4.1 Update Security card description with keyring, SQLCipher, Cloud KMS

## 5. Update docs/architecture/project-structure.md

- [x] 5.1 Update cli/security row to include all subcommands
- [x] 5.2 Add cli/p2p row with all P2P commands
- [x] 5.3 Add lifecycle, keyring, sandbox, dbmigrate rows to Infrastructure table
- [x] 5.4 Update security row to mention KMS providers
- [x] 5.5 Update Top-Level Layout skills line (remove "30")
- [x] 5.6 Update skills section description (removal explanation)

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./...` passes
