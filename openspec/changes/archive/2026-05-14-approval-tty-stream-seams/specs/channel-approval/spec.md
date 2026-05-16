## MODIFIED Requirements

### Requirement: TTY approval fallback behavior
The `TTYProvider.RequestApproval` SHALL return `(ApprovalResponse{}, error)` when stdin is not a terminal. When stdin is a terminal, it SHALL prompt with `[y/a/N]` where `a` means "always allow".

#### Scenario: TTY prompt uses deterministic streams in tests
- **WHEN** `TTYProvider.RequestApproval` is exercised in tests with injected terminal state and streams
- **THEN** the approval banner, optional summary, and `[y/a/N]` prompt SHALL be capturable without replacing process-global stdin or stderr
