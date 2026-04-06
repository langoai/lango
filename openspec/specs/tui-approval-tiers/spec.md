## ADDED Requirements

### Requirement: Two-tier approval classification
The system SHALL classify each approval request into one of two display tiers based on `SafetyLevel`, `Category`, and `Activity` fields on `ApprovalRequest`.

#### Scenario: Dangerous filesystem tool classified as Tier 2
- **WHEN** an approval request has `SafetyLevel: "dangerous"` and `Category: "filesystem"`
- **THEN** `ClassifyTier()` returns `TierFullscreen`

#### Scenario: Dangerous exec tool classified as Tier 2
- **WHEN** an approval request has `SafetyLevel: "dangerous"` and `Activity: "execute"`
- **THEN** `ClassifyTier()` returns `TierFullscreen`

#### Scenario: Moderate read tool classified as Tier 1
- **WHEN** an approval request has `SafetyLevel: "moderate"` and `Activity: "read"`
- **THEN** `ClassifyTier()` returns `TierInline`

### Requirement: ApprovalViewModel
An `ApprovalViewModel` struct SHALL bridge `ApprovalRequest` data to TUI rendering, carrying the computed `DisplayTier`, `RiskIndicator`, and optional `DiffContent`.

#### Scenario: ViewModel created from request
- **WHEN** the TUI approval provider receives an `ApprovalRequest` with populated SafetyLevel/Category/Activity
- **THEN** an `ApprovalViewModel` is created with the correct tier and risk indicator

### Requirement: Inline approval strip (Tier 1)
Tier 1 approvals SHALL render as a single-line compact strip below the transcript showing tool name, summary, and action keys.

#### Scenario: Inline strip display
- **WHEN** a Tier 1 approval is active
- **THEN** the strip renders: `[tool_name] summary... [a]llow [s]ession [d]eny`

#### Scenario: Inline strip key handling
- **WHEN** user presses `a` on an inline strip
- **THEN** the approval is granted with `AlwaysAllow: false`

### Requirement: Fullscreen approval dialog (Tier 2)
Tier 2 approvals SHALL render as a fullscreen overlay with risk badge, parameter display, diff preview, and action bar.

#### Scenario: Dialog display for file edit
- **WHEN** a Tier 2 approval for `fs_edit` is active
- **THEN** the dialog shows risk badge, tool name, summary, parameters, unified diff preview, and action bar

#### Scenario: Diff toggle
- **WHEN** user presses `t` in the approval dialog
- **THEN** the diff view toggles between unified and split format

#### Scenario: Diff viewport scrolling
- **WHEN** user presses `↑`/`↓` in the approval dialog
- **THEN** the diff viewport scrolls accordingly

#### Scenario: Diff generation with truncation
- **WHEN** the target file exceeds 500 lines
- **THEN** the diff is truncated with a `... (truncated)` marker

#### Scenario: Diff scroll and mode reset on new request
- **WHEN** a new ApprovalRequestMsg arrives
- **THEN** dialogScrollOffset resets to 0 and dialogSplitMode resets to false

#### Scenario: fs_write diff shows overwrite-aware header
- **WHEN** a Tier 2 approval for `fs_write` is displayed
- **THEN** the diff header shows `--- path (current or new)` / `+++ path (proposed)` to indicate possible overwrite

### Requirement: Risk indicator computation
`ComputeRisk()` SHALL return a `RiskIndicator` with level and label based on safety level and category.

#### Scenario: Dangerous exec shows critical risk
- **WHEN** `SafetyLevel: "dangerous"` and `Activity: "execute"`
- **THEN** risk indicator has `Level: "critical"` and `Label: "Executes arbitrary code"`

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
