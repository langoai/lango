## ADDED Requirements

### Requirement: 8 Context Layer Architecture
The system SHALL organize context into 8 distinct layers for retrieval-augmented generation.

#### Scenario: Context layer definitions
- **WHEN** the context retriever is initialized
- **THEN** the system SHALL recognize 8 layers: Tool Registry, User Knowledge, Skill Patterns, External Knowledge, Agent Learnings, Runtime Context, Observations, Reflections

### Requirement: Context Retrieval
The system SHALL search requested context layers and return relevant items.

#### Scenario: Retrieve from all default layers
- **WHEN** `Retrieve` is called with a query and no explicit layer list
- **THEN** the system SHALL search User Knowledge, Skill Patterns, External Knowledge, and Agent Learnings layers
- **AND** Tool Registry and Runtime Context SHALL be handled elsewhere

#### Scenario: Retrieve from specific layers
- **WHEN** `Retrieve` is called with an explicit list of layers
- **THEN** the system SHALL search only the specified layers

#### Scenario: Per-layer result limit
- **WHEN** retrieving context
- **THEN** each layer SHALL return at most `maxPerLayer` items (configurable, default 5)

#### Scenario: Empty query handling
- **WHEN** the query yields no keywords after extraction
- **THEN** the system SHALL return an empty result without querying any layer

#### Scenario: Layer retrieval error
- **WHEN** a layer query fails with an error
- **THEN** the system SHALL log a warning and continue with remaining layers

### Requirement: Keyword Extraction
The system SHALL extract meaningful keywords from user queries for search.

#### Scenario: Stop word filtering
- **WHEN** extracting keywords from a query
- **THEN** the system SHALL remove common English stop words (the, a, is, are, etc.)

#### Scenario: Short word filtering
- **WHEN** extracting keywords
- **THEN** the system SHALL remove words shorter than 2 characters
- **AND** SHALL preserve 2-character technical terms (e.g., "Go", "CI", "DB")

#### Scenario: Punctuation removal
- **WHEN** extracting keywords
- **THEN** the system SHALL strip punctuation from word boundaries

### Requirement: Prompt Assembly
The system SHALL assemble an augmented system prompt from base prompt, retrieved context, and observational memory.

#### Scenario: No context retrieved
- **WHEN** `AssemblePrompt` is called with no retrieved items and no observations
- **THEN** the system SHALL return the base prompt unchanged

#### Scenario: Context sections
- **WHEN** `AssemblePrompt` is called with retrieved items
- **THEN** the system SHALL append markdown sections for each layer with items:
  - "User Knowledge" for user knowledge items
  - "Known Solutions" for agent learnings
  - "Available Skills" for skill patterns
  - "External References" for external knowledge

#### Scenario: Observation memory section
- **WHEN** `AssemblePrompt` is called with observations or reflections
- **THEN** the system SHALL append a "Conversation Memory" section after knowledge sections
- **AND** reflections SHALL appear before observations within that section

### Requirement: Context-Aware Model Adapter
The system SHALL wrap the ADK model adapter to transparently inject retrieved context and observational memory.

#### Scenario: System prompt augmentation
- **WHEN** `GenerateContent` is called on the context-aware adapter
- **THEN** the system SHALL extract the user's latest message as query
- **AND** retrieve relevant context from all layers
- **AND** retrieve observations and reflections for the current session
- **AND** augment the system instruction with assembled context including observations
- **AND** forward the modified request to the underlying model adapter
