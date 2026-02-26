## ADDED Requirements

### Requirement: Observer Agent
The system SHALL provide an Observer agent that generates compressed observation notes from conversation history.

#### Scenario: Observation trigger on token threshold
- **WHEN** un-observed messages in a session exceed `messageTokenThreshold` tokens
- **THEN** the Observer SHALL generate a compressed observation note summarizing the un-observed messages
- **AND** store the observation with its token count and source message index range

#### Scenario: Observation content quality
- **WHEN** the Observer generates an observation
- **THEN** the observation SHALL capture: key decisions made, user intent and goals, important facts and context, task progress and outcomes
- **AND** the observation SHALL NOT include verbatim tool output or redundant detail

#### Scenario: Observer uses configurable LLM
- **WHEN** the Observer generates an observation
- **THEN** it SHALL use the provider and model specified in `observationalMemory.provider` and `observationalMemory.model`
- **AND** fall back to the primary agent's provider and model if not configured

#### Scenario: Observer failure graceful degradation
- **WHEN** the Observer LLM call fails
- **THEN** the system SHALL log the error and continue operating with raw message history
- **AND** retry observation on the next trigger cycle

### Requirement: Reflector Agent
The system SHALL provide a Reflector agent that condenses accumulated observations into reflections.

#### Scenario: Reflection trigger on observation token threshold
- **WHEN** accumulated observation tokens in a session exceed `observationTokenThreshold`
- **THEN** the Reflector SHALL generate a condensed reflection from all current observations
- **AND** store the reflection with its token count and generation number

#### Scenario: Observation replacement after reflection
- **WHEN** a reflection is successfully generated
- **THEN** the system SHALL delete the observations that were condensed into the reflection

#### Scenario: Multi-generation reflections
- **WHEN** reflections themselves accumulate beyond the threshold
- **THEN** the Reflector SHALL generate a higher-generation reflection condensing previous reflections
- **AND** increment the generation counter

### Requirement: Async Observation Buffer
The system SHALL support asynchronous observation generation via background goroutines.

#### Scenario: Non-blocking observation
- **WHEN** the observation trigger fires
- **THEN** the system SHALL send a signal to the background observer goroutine
- **AND** return immediately without blocking the message processing flow

#### Scenario: Goroutine lifecycle management
- **WHEN** the application starts with OM enabled
- **THEN** the system SHALL start the observer buffer goroutine
- **AND** register it with the application's WaitGroup for graceful shutdown

#### Scenario: Graceful shutdown
- **WHEN** the application receives a shutdown signal
- **THEN** the observer buffer SHALL stop accepting new signals
- **AND** complete any in-progress observation before exiting

### Requirement: Context Assembly with Observations
The system SHALL include observations and reflections in the context sent to the LLM. The session key for memory retrieval SHALL be resolved at call time from the request context via `session.SessionKeyFromContext(ctx)`, not from a field set at initialization.

#### Scenario: Context ordering
- **WHEN** assembling the augmented system prompt
- **THEN** the system SHALL order context as: base prompt, knowledge RAG, reflections, observations, recent messages

#### Scenario: No observations available
- **WHEN** a session has no observations or reflections
- **THEN** the context assembly SHALL behave identically to the current system (no change)

#### Scenario: Observation formatting
- **WHEN** observations are included in the prompt
- **THEN** they SHALL be formatted under a "## Conversation Memory" section
- **AND** reflections SHALL appear before observations within that section

#### Scenario: Session key resolved from request context
- **WHEN** `GenerateContent` is called with a context containing a session key (set by gateway or channel adapter)
- **THEN** the adapter SHALL extract the session key via `session.SessionKeyFromContext(ctx)`
- **AND** use it for memory retrieval, RAG session filtering, and runtime context updates

#### Scenario: No session key in context skips memory
- **WHEN** `GenerateContent` is called with a context that has no session key
- **THEN** the adapter SHALL skip memory retrieval entirely
- **AND** the LLM response SHALL not include a Conversation Memory section

#### Scenario: Session key propagated to RuntimeContextAdapter
- **WHEN** a session key is resolved from context
- **THEN** the adapter SHALL call `RuntimeContextAdapter.SetSession(sessionKey)`
- **AND** the runtime context SHALL reflect the correct session key and derived channel type

#### Scenario: Session key propagated to RAG retrieval
- **WHEN** a session key is resolved from context and RAG is enabled
- **THEN** the adapter SHALL pass the session key to RAG/GraphRAG retrieval options
- **AND** results SHALL be filtered by the session key

### Requirement: OM Configuration
The system SHALL support configuration of Observational Memory parameters.

#### Scenario: Configuration structure
- **WHEN** configuring OM
- **THEN** the config SHALL include: enabled (bool), provider (string), model (string), messageTokenThreshold (int), observationTokenThreshold (int), maxMessageTokenBudget (int)

#### Scenario: Default disabled
- **WHEN** no OM configuration is provided
- **THEN** the system SHALL default to disabled with zero impact on existing behavior

#### Scenario: Threshold defaults
- **WHEN** OM is enabled without explicit thresholds
- **THEN** messageTokenThreshold SHALL default to 1000
- **AND** observationTokenThreshold SHALL default to 2000
- **AND** maxMessageTokenBudget SHALL default to 8000

### Requirement: Delete reflections by session
The memory Store SHALL provide a `DeleteReflectionsBySession(ctx, sessionKey)` method that deletes all reflections for a given session key. The method SHALL follow the same pattern as `DeleteObservationsBySession`. The method SHALL return nil when the session has no reflections (no-op delete).

#### Scenario: Delete all reflections for a session
- **WHEN** `DeleteReflectionsBySession` is called with a session key that has reflections
- **THEN** all reflections for that session are deleted and reflections in other sessions are unaffected

#### Scenario: Delete from empty session
- **WHEN** `DeleteReflectionsBySession` is called with a session key that has no reflections
- **THEN** the method returns nil without error

### Requirement: Memory buffer compaction
The observational memory buffer SHALL support a compaction callback that deletes observed messages and replaces them with a summary after successful observation.

#### Scenario: Compaction after observation
- **WHEN** the memory buffer processes messages and generates an observation
- **THEN** the buffer SHALL invoke the compactor function to replace observed messages with a summary

#### Scenario: Compactor not set
- **WHEN** the buffer processes messages but no compactor is configured
- **THEN** the buffer SHALL skip compaction and retain all original messages

#### Scenario: Index reset after compaction
- **WHEN** compaction deletes messages up to index N and inserts a summary
- **THEN** the buffer SHALL reset its lastObserved index to 0 since message indices have shifted

### Requirement: SetCompactor configuration
The memory buffer SHALL provide a SetCompactor() method to configure the compaction callback at runtime.

#### Scenario: Wire compactor during app initialization
- **WHEN** the app wires memory buffer with session store
- **THEN** SetCompactor SHALL be called with EntStore.CompactMessages as the compaction function

### Requirement: Session store CompactMessages
The session store SHALL provide a CompactMessages(key, upToIndex, summary) method that atomically deletes messages up to the given index and inserts a summary message.

#### Scenario: Compact messages
- **WHEN** CompactMessages is called with key "session:123", upToIndex 5, and summary "User discussed weather"
- **THEN** messages at indices 0-5 SHALL be deleted and a new message with the summary SHALL be inserted with an early timestamp

#### Scenario: Compact with no messages to delete
- **WHEN** CompactMessages is called with upToIndex 0
- **THEN** the operation SHALL still insert the summary message

### Requirement: Recent reflections retrieval
The memory Store SHALL provide a `ListRecentReflections` method that returns the N most recent reflections ordered by created_at ascending (chronological) for a given session.

#### Scenario: Retrieve limited recent reflections
- **WHEN** ListRecentReflections is called with limit N
- **THEN** the N most recent reflections SHALL be returned in chronological order (oldest first)

### Requirement: Recent observations retrieval
The memory Store SHALL provide a `ListRecentObservations` method that returns the N most recent observations ordered by created_at ascending (chronological) for a given session.

#### Scenario: Retrieve limited recent observations
- **WHEN** ListRecentObservations is called with limit N
- **THEN** the N most recent observations SHALL be returned in chronological order (oldest first)

### Requirement: Configurable memory context limits
The ObservationalMemoryConfig SHALL support `MaxReflectionsInContext` (default: 5) and `MaxObservationsInContext` (default: 20) fields to limit the number of reflections and observations injected into the LLM context. Zero means unlimited.

#### Scenario: Default limits applied
- **WHEN** maxReflectionsInContext and maxObservationsInContext are not configured (zero)
- **THEN** defaults of 5 reflections and 20 observations SHALL be applied by the wiring layer

#### Scenario: Custom limits applied
- **WHEN** maxReflectionsInContext is set to 3 and maxObservationsInContext is set to 10
- **THEN** only the 3 most recent reflections and 10 most recent observations SHALL be injected into context

### Requirement: Memory token budgeting in context assembly
The `ContextAwareModelAdapter` SHALL enforce a token budget when assembling the memory section into the system prompt. Reflections SHALL be included first (higher information density), then observations fill the remaining budget.

#### Scenario: Default memory token budget
- **WHEN** no explicit budget is configured via `WithMemoryTokenBudget`
- **THEN** the default budget SHALL be 4000 tokens

#### Scenario: Reflections exceed budget
- **WHEN** reflections alone exceed the token budget
- **THEN** the system SHALL include reflections up to the budget limit and skip all observations

#### Scenario: Budget shared between reflections and observations
- **WHEN** reflections use part of the budget
- **THEN** observations SHALL fill the remaining budget, stopping when the next observation would exceed it

#### Scenario: Custom budget via WithMemoryTokenBudget
- **WHEN** `WithMemoryTokenBudget(budget)` is called with a positive value
- **THEN** the adapter SHALL use that budget instead of the default 4000

### Requirement: Auto meta-reflection on accumulation
The `memory.Buffer` SHALL automatically trigger meta-reflection when the number of reflections in a session exceeds a configurable consolidation threshold.

#### Scenario: Default consolidation threshold
- **WHEN** no explicit threshold is configured
- **THEN** the default threshold SHALL be 5 reflections

#### Scenario: Meta-reflection triggered
- **WHEN** `process()` completes and the session has >= threshold reflections
- **THEN** `ReflectOnReflections` SHALL be called to consolidate them

#### Scenario: Meta-reflection failure is non-fatal
- **WHEN** `ReflectOnReflections` returns an error
- **THEN** the system SHALL log the error and continue normal operation

### Requirement: Buffer drops logged at warn level with counters
EmbeddingBuffer and GraphBuffer SHALL log dropped requests at warn level (not debug) and track drop counts via atomic counters accessible through a DroppedCount() method.

#### Scenario: Queue full logs warning
- **WHEN** a buffer's queue is full and a new request arrives
- **THEN** the request SHALL be dropped with a warn-level log entry including the request ID

#### Scenario: Drop counter increments
- **WHEN** a buffer drops a request
- **THEN** the atomic drop counter SHALL increment and be readable via DroppedCount()
