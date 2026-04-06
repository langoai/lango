### Requirement: Delegation events displayed in transcript
The system SHALL display agent-to-agent delegation events as `itemDelegation` transcript items showing the source agent, target agent, and optional reason.

#### Scenario: Outward delegation displayed
- **WHEN** an `OnDelegation` callback fires with from="operator", to="librarian", reason="search needed"
- **THEN** the transcript SHALL display a delegation item with "🔀 operator → librarian" and the reason text

#### Scenario: Orchestrator return hop excluded from counter
- **WHEN** a delegation targets "lango-orchestrator"
- **THEN** the delegation counter SHALL NOT increment, but the active agent label SHALL update

### Requirement: Recovery events displayed in transcript
The system SHALL display recovery decision events as `itemRecovery` transcript items showing the action type, cause class, attempt number, and backoff duration.

#### Scenario: Retry recovery displayed
- **WHEN** a `RecoveryDecisionEvent` fires with action="retry", causeClass="rate_limit", attempt=2, backoff=2s
- **THEN** the transcript SHALL display "🔄 Retry #2 (rate_limit) 2s backoff"

#### Scenario: Reroute recovery displayed
- **WHEN** a `RecoveryDecisionEvent` fires with action="retry_with_hint"
- **THEN** the transcript SHALL display "Reroute" as the action label

### Requirement: Budget warnings displayed in transcript
The system SHALL display delegation budget warnings as warning-toned status items when the budget threshold is crossed.

#### Scenario: Budget warning at 80%
- **WHEN** an `OnBudgetWarning` callback fires with used=12, max=15
- **THEN** the transcript SHALL display "⚠ Delegation budget: 12/15 (80%)" as a warning status

### Requirement: Per-turn token usage summary
The system SHALL display a token usage summary after each turn's assistant response, showing input, output, total, and cache token counts.

#### Scenario: Token summary after response
- **WHEN** a turn completes with accumulated token usage
- **THEN** the token summary SHALL appear AFTER the assistant response in the transcript

#### Scenario: No summary for zero tokens
- **WHEN** a turn completes with zero accumulated tokens
- **THEN** no token summary SHALL be displayed

### Requirement: Runtime messages reach chat from any page
The system SHALL forward DelegationMsg, BudgetWarningMsg, RecoveryMsg, and DoneMsg to the chat child regardless of which cockpit page is active.

#### Scenario: Budget warning while on Settings page
- **WHEN** the user is viewing the Settings page and a BudgetWarningMsg arrives
- **THEN** the message SHALL be forwarded to the chat child model

### Requirement: Delegation reason text sanitized
The system SHALL sanitize delegation reason text with `ansi.Strip` before rendering to prevent terminal control injection.

#### Scenario: ANSI escape in reason text
- **WHEN** a delegation reason contains ANSI escape sequences
- **THEN** the escape sequences SHALL be stripped before display
