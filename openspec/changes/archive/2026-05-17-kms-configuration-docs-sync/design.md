# Design

## Approach

This is a documentation synchronization change with a test guard. The runtime source of truth remains:

- `internal/config/types_security.go` for profile-backed `security.kms.*` fields.
- `internal/bootstrap/kms_env.go` for environment-backed bootstrap KMS options.

`README.md` and `docs/configuration.md` should expose the same operator-facing KMS settings from config structs, while clarifying that profile config is not available until after encrypted profile bootstrap credentials are acquired.

## Scope Control

The change intentionally updates only the public configuration references and one docs guard test. It avoids broad doc reshaping or additional planning artifacts beyond the OpenSpec delta needed to archive the work.

## Verification

- A focused `internal/testutil` test will fail before the docs update and pass after it.
- OpenSpec validation will cover the new docs-sync requirement.
- Full Go build and test will run before commit.
