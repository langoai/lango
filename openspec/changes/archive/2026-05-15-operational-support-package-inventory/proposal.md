## Why

The repository ships operational-support packages under `internal/alerting`, `internal/approvalflow`, `internal/archtest`, and `internal/dbopen`, but those packages are still omitted from the public inventory docs.

## What Changes

- add the missing operational-support package rows to the README internal package tree
- add matching package rows to `docs/architecture/project-structure.md`
- add an executable guard that requires those package rows and their current responsibilities
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for alerting, approval-flow, architecture-test, and DB-open support packages
- better discoverability of operational and production-readiness internals
- stronger regression protection against future omissions or vague wording
