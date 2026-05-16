## 1. Implementation

- [x] 1.1 Route `lango workflow list` table and empty-state output through `cmd.OutOrStdout()`
- [x] 1.2 Route `lango workflow status` detail output through `cmd.OutOrStdout()`
- [x] 1.3 Route `lango workflow history` table and empty-state output through `cmd.OutOrStdout()`
- [x] 1.4 Route `lango workflow cancel` success output through `cmd.OutOrStdout()`
- [x] 1.5 Add a deterministic cancel seam for command-level tests

## 2. Tests

- [x] 2.1 Add regression coverage for workflow list output capture
- [x] 2.2 Add regression coverage for workflow status output capture
- [x] 2.3 Add regression coverage for workflow history output capture
- [x] 2.4 Add regression coverage for workflow cancel confirmation capture

## 3. Downstream

- [x] 3.1 Update public workflow CLI docs for writer routing
- [x] 3.2 Update OpenSpec main spec with the writer-routing contract
- [x] 3.3 Add change proposal
- [x] 3.4 Add change design
- [x] 3.5 Add delta spec
