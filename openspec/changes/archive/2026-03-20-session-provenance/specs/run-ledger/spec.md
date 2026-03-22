## ADDED Requirements

### Requirement: Store Option Pattern
MemoryStore and EntStore constructors SHALL accept variadic `StoreOption` parameters via `WithAppendHook(func(JournalEvent))`. The `RunLedgerStore` interface SHALL NOT be modified.

#### Scenario: Backward compatible construction
- **WHEN** `NewMemoryStore()` or `NewEntStore(client)` is called without options
- **THEN** behavior is identical to pre-change behavior

#### Scenario: Append hook registration
- **WHEN** `NewMemoryStore(WithAppendHook(h))` is called
- **THEN** the hook `h` is called after each successful journal event append

#### Scenario: Hook runs outside lock
- **WHEN** an append hook reads from the same MemoryStore it is registered on
- **THEN** no deadlock occurs because the hook is invoked after the write lock is released
