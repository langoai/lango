## 1. Review Recent Refactors

- [x] 1.1 Read the latest four commits and map their behavior to existing OpenSpec capabilities
- [x] 1.2 Confirm current implementation details in the affected packages before writing specs

## 2. Write Delta Specs

- [x] 2.1 Add an `agent-memory` delta covering tool ownership, context-aware kind filtering, and runtime kind validation
- [x] 2.2 Add an `automation-agent-tools` delta covering shared automation interfaces and channel detection
- [x] 2.3 Add a `domain-tool-builders` delta covering extracted builders outside the economy package
- [x] 2.4 Add a `tool-catalog` delta aligning sentinel registration and app-owned builder exceptions
- [x] 2.5 Add a `parity-verification` delta covering extracted tool builder parity tests

## 3. Sync Main Specs

- [x] 3.1 Apply the delta specs to the corresponding files under `openspec/specs/`
- [x] 3.2 Verify the updated main specs match current code behavior and ownership boundaries

## 4. Finalize Documentation Change

- [x] 4.1 Confirm the change artifacts fully describe the recent four commits
- [x] 4.2 Leave the repository with synced main specs and a complete OpenSpec documentation trail
