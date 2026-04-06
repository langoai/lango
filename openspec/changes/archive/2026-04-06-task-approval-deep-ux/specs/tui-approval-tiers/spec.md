## MODIFIED Requirements

### Requirement: Approval view model
The ApprovalViewModel SHALL include a RuleExplanation string field populated by buildRuleExplanation() based on the request's safety level and category.

#### Scenario: ViewModel includes explanation
- **WHEN** NewViewModel is called with a dangerous filesystem request
- **THEN** RuleExplanation contains "filesystem" and "dangerous"

### Requirement: Approval dialog rendering
The approval dialog SHALL render a "Why: ..." explanation line between the summary and parameters sections. For critical-risk tools, the action bar SHALL show "Press again to confirm" when confirmPending is true.

#### Scenario: Explanation displayed
- **WHEN** an approval dialog is rendered with a non-empty RuleExplanation
- **THEN** the output contains "Why:" followed by the explanation

#### Scenario: Confirm prompt displayed
- **WHEN** confirmPending is true
- **THEN** the action bar shows "Press 'a' again to confirm" instead of normal keys

### Requirement: Approval strip rendering
The approval strip SHALL prepend "(destructive)" in red for critical-risk tools and show the confirm prompt when confirmPending is true.

#### Scenario: Destructive label shown
- **WHEN** a critical-risk approval is rendered in strip mode
- **THEN** "(destructive)" appears after the tool badge

#### Scenario: Non-critical no label
- **WHEN** a non-critical approval is rendered in strip mode
- **THEN** "(destructive)" does NOT appear
