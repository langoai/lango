## Why

The repository ships an execution and retrieval infrastructure stack under `internal/agentrt`, `internal/gatekeeper`, `internal/retrieval`, `internal/search`, `internal/turnrunner`, `internal/turntrace`, `internal/lineio`, and `internal/storeutil`, but those packages are still omitted or under-described in the public inventory docs.

## What Changes

- add the missing execution-retrieval infrastructure package rows to the README internal package tree
- add matching package rows to `docs/architecture/project-structure.md`
- add an executable guard that requires those package rows and their current responsibilities
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for the execution/retrieval infrastructure stack
- better discoverability of runtime coordination, sanitization, search, and trace internals
- stronger regression protection against future omissions or vague wording
