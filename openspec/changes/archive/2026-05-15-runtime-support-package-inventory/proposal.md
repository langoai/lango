## Why

The repository ships `internal/exportability`, `internal/knowledgeruntime`, `internal/receipts`, `internal/storagebroker`, `internal/streamx`, `internal/tooloutput`, and `internal/toolparam`, but the public inventory docs omit them from both the README internal tree and the architecture project-structure reference.

## What Changes

- add those runtime-support packages to the README internal package tree
- add matching package rows to `docs/architecture/project-structure.md`
- add an executable guard that requires those package rows and their current responsibilities
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for shipped runtime support packages
- better discoverability of receipt/runtime/storage/tooling internals
- stronger regression protection against future omissions or vague wording
