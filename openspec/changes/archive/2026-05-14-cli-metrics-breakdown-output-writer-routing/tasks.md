## 1. Implementation

- [x] 1.1 Make shared metrics JSON and tabwriter helpers writer-aware
- [x] 1.2 Route `sessions`, `tools`, `agents`, and `history` output through `cmd.OutOrStdout()`

## 2. Tests

- [x] 2.1 Add regression coverage for sessions table and JSON output capture
- [x] 2.2 Add regression coverage for tools/agents empty-state capture
- [x] 2.3 Add regression coverage for history table output capture

## 3. Downstream

- [x] 3.1 Update public metrics CLI docs for writer routing
- [x] 3.2 Update OpenSpec main spec with the breakdown writer-routing contract
- [x] 3.3 Add change proposal
- [x] 3.4 Add change design
- [x] 3.5 Add delta spec
