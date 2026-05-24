## 1. Tests

- [x] 1.1 Add a failing docs guard for public Cloud KMS configuration settings.
- [x] 1.2 Add a failing docs guard for the bootstrap fallback env override note.

## 2. Documentation

- [x] 2.1 Add a Cloud KMS subsection to `docs/configuration.md`.
- [x] 2.2 List all profile-backed `security.kms.*` fields currently exposed by config in README and `docs/configuration.md`.
- [x] 2.3 Explain `LANGO_KMS_FALLBACK_TO_LOCAL=false` for env-driven encrypted profile bootstrap.

## 3. Verification

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run the focused docs guard test.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent-driven review.
- [x] 3.5 Archive the OpenSpec change.
- [x] 3.6 Commit this scoped unit separately.
