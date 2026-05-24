## Why

Package-local regression tests for `internal/dbopen` exposed a real concurrency hazard: concurrent `OpenManaged` calls can crash inside Ent Atlas schema setup with `concurrent map writes`.

## What Changes

- serialize `client.Schema.Create` inside `internal/dbopen.OpenManaged`
- add a concurrent `OpenManaged` regression test that fails on panic or migration/open errors
- update public package inventory wording so `dbopen` no longer implies fully parallel-safe migration
- sync the main docs-only and test-coverage specs

## Impact

- removes a production-facing crash path during concurrent managed database opens
- keeps direct low-level DB-open regression coverage close to the package
- makes the architecture/package inventory more truthful about `dbopen` behavior
