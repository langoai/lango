## ADDED Requirements

### Requirement: RetrievalConfig Feedback field
The `RetrievalConfig` struct SHALL include a `Feedback bool` field that enables context injection observability. This field SHALL operate independently of `Enabled` and `Shadow` — feedback observability SHALL work regardless of whether the agentic retrieval coordinator is enabled.

#### Scenario: Feedback enabled without coordinator
- **WHEN** `retrieval.feedback` is `true` and `retrieval.enabled` is `false`
- **THEN** the `FeedbackProcessor` SHALL be subscribed to the event bus

#### Scenario: Feedback default
- **WHEN** no `retrieval.feedback` value is configured
- **THEN** feedback SHALL default to `false`

### Requirement: Event bus wiring on ContextAwareModelAdapter
The `ContextAwareModelAdapter` SHALL accept an event bus via `WithEventBus(*eventbus.Bus)`. The event bus SHALL be wired unconditionally when a `ContextAwareModelAdapter` exists, regardless of coordinator presence.

#### Scenario: Bus wired in knowledge branch
- **WHEN** knowledge system is enabled and ctxAdapter is created
- **THEN** `WithEventBus(eventBus)` SHALL be called on the adapter

#### Scenario: Bus wired in OM-only branch
- **WHEN** only observational memory is enabled (no knowledge) and ctxAdapter is created
- **THEN** `WithEventBus(eventBus)` SHALL be called on the adapter
