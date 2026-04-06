## MODIFIED Requirements

### Requirement: ChatModel uses composite sub-models for orthogonal state
ChatModel SHALL use composite sub-model types for CPR filtering (`cprFilter`), pending indicator (`pendingIndicator`), and approval workflow (`approvalState`). These SHALL replace primitive field groups, reducing the ChatModel struct from 23 fields to 18 fields.

#### Scenario: CPR state accessed via cpr field
- **WHEN** ChatModel processes terminal response sequences
- **THEN** it SHALL delegate to `m.cpr.Filter()`, `m.cpr.Flush()`, and `m.cpr.HandleTimeout()`

#### Scenario: Pending state accessed via pending field
- **WHEN** user submits input and waits for first content event
- **THEN** ChatModel SHALL call `m.pending.Activate()` on submit and `m.pending.Dismiss()` on first content event
- **AND** `m.pending.IsActive()` and `m.pending.Elapsed()` SHALL be used for rendering

#### Scenario: Approval state accessed via approval field
- **WHEN** an approval request arrives
- **THEN** ChatModel SHALL call `m.approval.Reset(&msg)` to initialize and `m.approval.Clear()` after response
- **AND** `m.approval.pending`, `m.approval.confirmPending` SHALL be used for rendering and key handling
