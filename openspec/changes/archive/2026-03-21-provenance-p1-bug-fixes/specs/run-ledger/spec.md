## ADDED Requirements

### Requirement: AppendHookSetter interface
Concrete store types (`MemoryStore`, `EntStore`) SHALL implement the `AppendHookSetter` interface with a `SetAppendHook(func(JournalEvent))` method for post-construction hook registration. This interface is NOT part of the `RunLedgerStore` contract.

#### Scenario: Post-construction hook registration
- **WHEN** `SetAppendHook` is called on a store after construction
- **THEN** the registered hook is invoked on subsequent `AppendJournalEvent` calls

#### Scenario: Hook chaining preserves existing hooks
- **WHEN** a store is created with `WithAppendHook(first)` and then `SetAppendHook(second)` is called
- **THEN** both `first` and `second` are invoked in order on each `AppendJournalEvent` call
