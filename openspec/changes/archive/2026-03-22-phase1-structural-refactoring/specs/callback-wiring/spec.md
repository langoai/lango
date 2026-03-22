## REMOVED Requirements

### Requirement: SetEmbedCallback on knowledge and memory stores
**Reason**: Replaced by EventBus `ContentSavedEvent` publish pattern. Stores now call `bus.Publish(ContentSavedEvent{...})` instead of invoking embed callbacks directly.
**Migration**: Call `store.SetEventBus(bus)` instead of `store.SetEmbedCallback(cb)`. Subscribe to `ContentSavedEvent` in wiring files.

### Requirement: SetGraphCallback on knowledge, memory, learning stores
**Reason**: Replaced by EventBus pattern. Graph wiring subscribes to `ContentSavedEvent` (filtered by `NeedsGraph`) and `TriplesExtractedEvent`.
**Migration**: Call `store.SetEventBus(bus)` or `engine.SetEventBus(bus)`. Graph routing is handled by `wiring_graph.go` EventBus subscriptions.

### Requirement: SetOnChangeCallback on reputation store
**Reason**: Replaced by EventBus `ReputationChangedEvent`.
**Migration**: Call `repStore.SetEventBus(bus)`. Subscribe to `ReputationChangedEvent` in `wiring_p2p.go`.

## ADDED Requirements

### Requirement: Stores accept EventBus via SetEventBus
Knowledge store, memory store, learning engines (GraphEngine, ConversationAnalyzer, SessionLearner), librarian ProactiveBuffer, and reputation store SHALL accept an `*eventbus.Bus` via `SetEventBus(bus)` method.

#### Scenario: Store publishes events when bus is set
- **WHEN** a store has a bus set via SetEventBus
- **AND** content is saved
- **THEN** the appropriate event is published on the bus

#### Scenario: Store silently skips publish when bus is nil
- **WHEN** a store has no bus set (nil)
- **AND** content is saved
- **THEN** no panic occurs and no event is published
