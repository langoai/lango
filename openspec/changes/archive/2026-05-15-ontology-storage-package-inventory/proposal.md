## Why

The repository ships core ontology and storage foundation packages under `internal/ontology`, `internal/sqlitedriver`, and `internal/storage`, but they remain omitted from the public inventory docs.

## What Changes

- add the missing ontology/storage package rows to the README internal package tree
- add matching package rows to `docs/architecture/project-structure.md`
- add an executable guard that requires those package rows and their current responsibilities
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for ontology and storage foundations
- better discoverability of governance, DB-open, and storage-facade internals
- stronger regression protection against future omissions or vague wording
