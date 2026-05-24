## MODIFIED Requirements

### Requirement: TTY approval fallback behavior
The `TTYProvider.RequestApproval` SHALL return `(ApprovalResponse{}, error)` when stdin is not a terminal. When stdin is a terminal, it SHALL prompt with `[y/a/N]` where `a` means "always allow".

#### Scenario: Terminal prompt EOF denies safely
- **WHEN** the TTY prompt reaches EOF before an approval answer is received
- **THEN** it SHALL return `ApprovalResponse{Approved: false, AlwaysAllow: false}`
- **AND** it SHALL NOT surface a read error
