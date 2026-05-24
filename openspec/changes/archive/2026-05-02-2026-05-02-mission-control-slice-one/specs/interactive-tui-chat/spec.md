## MODIFIED Requirements

### Requirement: Interactive TUI chat on bare invocation
Running `lango` without arguments SHALL no longer claim direct chat as the default interactive TUI surface. Focused interactive chat remains available through `lango chat`, and the chat model continues to power that direct surface.

#### Scenario: Bare `lango` no longer launches direct chat
- **WHEN** the user runs `lango` with no arguments on an interactive terminal
- **THEN** the application SHALL open Mission Control instead of launching the focused chat surface directly
- **AND** this delta SHALL replace the earlier bare-`lango` chat-first assumption for the default TUI entrypoint

#### Scenario: `lango chat` still launches focused chat
- **WHEN** the user runs `lango chat`
- **THEN** an interactive TUI chat session SHALL start

## ADDED Requirements

### Requirement: ChatModel cooperates with cockpit-owned pending approvals
When the chat model is mounted inside cockpit, pending approval ownership SHALL be shared with the cockpit-level pending approval owner rather than being duplicated inside chat. Standalone `lango chat` approval behavior remains unchanged.

#### Scenario: Cockpit chat reads the shared pending request
- **WHEN** ChatModel is running inside cockpit and a pending approval exists
- **THEN** chat rendering and key handling SHALL read the same pending approval state owned by cockpit
- **AND** chat SHALL NOT require a second independent pending approval copy

#### Scenario: Standalone chat keeps direct approval ownership
- **WHEN** ChatModel runs through `lango chat` outside cockpit
- **THEN** it SHALL continue to own and resolve approvals through its direct interactive path

## MODIFIED Requirements

### Requirement: Learning suggestion rendering in TUI
Slice 1 SHALL NOT require the chat transcript surface to present learning suggestions as inline approvals that persist learning directly. Mission Control becomes the required projection surface for those suggestions as actionable proposed missions, and chat may remain informational if it renders them at all.

#### Scenario: Mission Control owns learning suggestion proposal semantics
- **WHEN** a `LearningSuggestionEvent` occurs during a cockpit session
- **THEN** the Slice 1 requirement SHALL be satisfied by Mission Control projecting it as a proposed mission
- **AND** this delta SHALL replace the earlier requirement that chat itself present the suggestion as an inline approval with direct persistence semantics
- **AND** chat SHALL NOT be required to persist the suggestion through the approval pipeline on its own
