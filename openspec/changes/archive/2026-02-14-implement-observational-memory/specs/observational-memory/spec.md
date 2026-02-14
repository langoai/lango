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
The system SHALL include observations and reflections in the context sent to the LLM.

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
