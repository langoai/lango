## ADDED Requirements

### Requirement: Approval rule explanation
The approval dialog and strip SHALL display a "Why" explanation describing why the tool requires approval, based on safety level and category.

#### Scenario: Dangerous filesystem tool
- **WHEN** an approval is requested for a dangerous filesystem tool
- **THEN** the explanation reads "This tool modifies the filesystem and is classified as dangerous."

#### Scenario: Dangerous automation tool
- **WHEN** an approval is requested for a dangerous automation tool
- **THEN** the explanation reads "This tool executes arbitrary code and is classified as dangerous."

#### Scenario: Moderate tool
- **WHEN** an approval is requested for a moderate-risk tool
- **THEN** the explanation reads "This tool creates or modifies resources (moderate risk)."

#### Scenario: Default explanation
- **WHEN** an approval is requested for a safe-level tool
- **THEN** the explanation reads "This tool requires approval under the current approval policy."

### Requirement: Double-press confirmation for critical approvals
The approval flow SHALL require a double-press of `a` or `s` for critical-risk tools. The first press shows a "Press again to confirm" prompt. The second press of the SAME key within 3 seconds executes the action.

#### Scenario: Critical tool first press of allow
- **WHEN** user presses `a` on a critical-risk approval for the first time
- **THEN** confirmPending is set with action="a" and no approval is sent

#### Scenario: Critical tool second press of allow
- **WHEN** user presses `a` again within 3 seconds while action="a" is pending
- **THEN** the tool is approved (one-time allow)

#### Scenario: Critical tool session allow tracks action
- **WHEN** user presses `s` on a critical-risk approval
- **THEN** confirmPending is set with action="s", and second `s` press sends session grant

#### Scenario: Mismatched second press
- **WHEN** user presses `s` first then `a` second on critical-risk
- **THEN** the pending state is reset and `a` starts a new confirm cycle (not executed)

#### Scenario: Confirm timeout
- **WHEN** more than 3 seconds elapse after the first press
- **THEN** confirmPending is cleared

#### Scenario: Non-critical immediate approval
- **WHEN** user presses `a` on a non-critical approval
- **THEN** the tool is approved immediately without double-press

### Requirement: Destructive label on critical strip
The approval strip SHALL display "(destructive)" in red after the tool badge for critical-risk tools.

#### Scenario: Critical strip label
- **WHEN** a critical-risk approval is shown in strip mode
- **THEN** the strip includes "(destructive)" text

### Requirement: ApprovalRequestMsg auto-switches to chat
The cockpit SHALL automatically switch to the Chat page when an ApprovalRequestMsg arrives while a non-chat page is active, ensuring the user can see and respond to the prompt.

#### Scenario: Approval on non-chat page
- **WHEN** an ApprovalRequestMsg arrives while Tasks page is active
- **THEN** the cockpit switches to Chat page and forwards the message to the chat child
