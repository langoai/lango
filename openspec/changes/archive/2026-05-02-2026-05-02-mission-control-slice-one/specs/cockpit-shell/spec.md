## ADDED Requirements

### Requirement: Cockpit root model orchestrates Mission Control as the default first surface
The cockpit root model SHALL treat Mission Control as the default active page for bare `lango` launches while continuing to host the existing detail pages. The sidebar remains secondary navigation and SHALL NOT be required for the user to understand the current system state on first render.

#### Scenario: Default active page is Mission Control
- **WHEN** cockpit is created for the default `lango` surface
- **THEN** `activePage` SHALL initialize to `PageMissionControl`

#### Scenario: Chat remains a detail page inside cockpit
- **WHEN** the user navigates from Mission Control to Chat inside cockpit
- **THEN** the chat child surface SHALL remain reachable as a detail route
- **AND** `lango chat` SHALL still bypass cockpit Mission Control entirely

### Requirement: Pending approvals use a cockpit-owned shared response path
The cockpit shell SHALL own the latest pending approval request for the session through a shared pending approval registry. Approval requests SHALL no longer require an unconditional page switch to Chat in order to remain visible or resolvable.

#### Scenario: Approval request does not force Chat takeover
- **WHEN** a pending `ApprovalRequestMsg` arrives while Mission Control is active
- **THEN** cockpit SHALL register the request with the shared pending approval owner
- **AND** Mission Control MAY remain visible while the live decision is rendered

#### Scenario: Shared pending request resolves once
- **WHEN** a cockpit surface resolves the pending approval
- **THEN** the cockpit-owned registry SHALL write exactly one response to the original approval response channel
- **AND** the pending request SHALL clear for all cockpit surfaces

### Requirement: Mission Control shared subscriptions are scoped to cockpit lifetime
Cockpit SHALL own the EventBus subscriptions and shared buffers needed for Mission Control for the lifetime of the TUI session. Page activation and deactivation SHALL control rendering and refresh ticks only, not EventBus subscription teardown.

#### Scenario: Page switch preserves shared state
- **WHEN** the user switches away from Mission Control and later returns
- **THEN** the shared mission activity, learning suggestion, and pending approval state SHALL still be available

#### Scenario: No page-lifetime unsubscribe contract
- **WHEN** Mission Control deactivates as a page
- **THEN** the cockpit SHALL stop page-local refresh behavior only
- **AND** the change SHALL NOT require EventBus `Unsubscribe()` support
