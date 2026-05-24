## 1. Implementation

- [x] 1.1 Route `lango workflow run` validation and schedule guidance output through `cmd.OutOrStdout()`
- [x] 1.2 Route direct execution status output through `cmd.OutOrStdout()`
- [x] 1.3 Replace stdout-swapping workflow run schedule test with writer-based capture

## 2. Tests

- [x] 2.1 Keep regression coverage for schedule-not-implemented guidance on the command writer

## 3. Downstream

- [x] 3.1 Update public workflow CLI docs for writer routing
- [x] 3.2 Update OpenSpec main spec with the writer-routing contract
- [x] 3.3 Add change proposal
- [x] 3.4 Add change design
- [x] 3.5 Add delta spec
