## 1. Implementation

- [x] 1.1 Route `lango account session create` table output through `cmd.OutOrStdout()`
- [x] 1.2 Route `lango account session create --output json` through `cmd.OutOrStdout()`
- [x] 1.3 Route `lango account session list` table, JSON, and empty-state output through `cmd.OutOrStdout()`
- [x] 1.4 Route `lango account session revoke` success output through `cmd.OutOrStdout()`
- [x] 1.5 Add deterministic session seams for command-level tests

## 2. Tests

- [x] 2.1 Add regression coverage for session create table output capture
- [x] 2.2 Add regression coverage for session create JSON output capture
- [x] 2.3 Add regression coverage for session list table/JSON/empty-state output capture
- [x] 2.4 Add regression coverage for session revoke single/all success output capture

## 3. Downstream

- [x] 3.1 Update public smart account CLI docs for writer routing
- [x] 3.2 Update OpenSpec main spec with the writer-routing contract
- [x] 3.3 Add change proposal
- [x] 3.4 Add change design
- [x] 3.5 Add delta spec
