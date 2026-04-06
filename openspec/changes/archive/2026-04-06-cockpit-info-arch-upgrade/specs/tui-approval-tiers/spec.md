## MODIFIED Requirements

### Requirement: Approval dialog state owned by approvalState struct
The approval dialog scroll position and split mode SHALL be owned by the `approvalState` struct instead of package-level globals. `renderApprovalDialog` SHALL accept `scrollOffset int` and `splitMode bool` as explicit parameters. `handleApprovalDialogKey` SHALL accept `*approvalState` and mutate scroll/split via `ScrollDiff()` and `ToggleSplit()` methods.

#### Scenario: Dialog scroll via approvalState
- **WHEN** user scrolls the approval dialog with up/down keys
- **THEN** `handleApprovalDialogKey` SHALL call `state.ScrollDiff()` to adjust the offset
- **AND** `renderApprovalDialog` SHALL receive the updated `scrollOffset` as a parameter

#### Scenario: Dialog split toggle via approvalState
- **WHEN** user presses the split toggle key in the approval dialog
- **THEN** `handleApprovalDialogKey` SHALL call `state.ToggleSplit()`
- **AND** `renderApprovalDialog` SHALL receive the updated `splitMode` as a parameter

#### Scenario: Dialog state reset on new approval
- **WHEN** a new `ApprovalRequestMsg` arrives
- **THEN** `approvalState.Reset()` SHALL set scrollOffset to 0 and splitMode to false

#### Scenario: No package-level globals for dialog state
- **WHEN** the chat package is compiled
- **THEN** there SHALL be no package-level `var` declarations for `dialogScrollOffset` or `dialogSplitMode`
