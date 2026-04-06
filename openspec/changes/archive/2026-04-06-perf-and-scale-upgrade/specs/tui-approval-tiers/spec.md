## MODIFIED Requirements

### Requirement: Approval dialog diff rendering cached
The approval dialog SHALL cache styled diff lines in a diffLineCache struct keyed by content, width, and splitMode. Scrolling SHALL window over the cached styled line slice without re-styling. The cache SHALL be cleared on Reset() (new approval), ToggleSplit(), and Clear() (approval end).

#### Scenario: Diff lines cached on first render
- **WHEN** renderApprovalDialog renders diff content for the first time (or after cache invalidation)
- **THEN** all diff lines SHALL be styled and stored in state.diffCache.lines

#### Scenario: Scroll uses cached lines via windowing
- **WHEN** the user scrolls the approval dialog
- **THEN** the render function SHALL slice state.diffCache.lines[scrollOffset:scrollOffset+viewHeight] without re-styling

#### Scenario: Cache invalidated on new approval
- **WHEN** approvalState.Reset() is called with a new approval request
- **THEN** diffCache SHALL be cleared to zero value

#### Scenario: Cache invalidated on split toggle
- **WHEN** approvalState.ToggleSplit() is called
- **THEN** diffCache.lines SHALL be set to nil

#### Scenario: Cache released on approval end
- **WHEN** approvalState.Clear() is called after approval response
- **THEN** diffCache SHALL be cleared to release memory

### Requirement: renderApproval accepts approvalState
The renderApproval dispatcher SHALL accept `*approvalState` directly instead of individual scrollOffset, splitMode, and confirmPending parameters. renderApprovalDialog SHALL read scroll/split/cache state from the approvalState struct.

#### Scenario: renderApproval signature
- **WHEN** renderApproval is called from RenderParts or recalcLayout
- **THEN** it SHALL receive `(msg *ApprovalRequestMsg, state *approvalState, width, height int)` parameters
