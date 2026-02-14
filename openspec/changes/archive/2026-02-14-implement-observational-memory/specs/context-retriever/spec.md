## MODIFIED Requirements

### Requirement: 6 Context Layer Architecture
The system SHALL organize context into 8 distinct layers for retrieval-augmented generation.

#### Scenario: Context layer definitions
- **WHEN** the context retriever is initialized
- **THEN** the system SHALL recognize 8 layers: Tool Registry, User Knowledge, Skill Patterns, External Knowledge, Agent Learnings, Runtime Context, Observations, Reflections

## MODIFIED Requirements

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

## MODIFIED Requirements

### Requirement: Context-Aware Model Adapter
The system SHALL wrap the ADK model adapter to transparently inject retrieved context and observational memory.

#### Scenario: System prompt augmentation
- **WHEN** `GenerateContent` is called on the context-aware adapter
- **THEN** the system SHALL extract the user's latest message as query
- **AND** retrieve relevant context from all layers
- **AND** retrieve observations and reflections for the current session
- **AND** augment the system instruction with assembled context including observations
- **AND** forward the modified request to the underlying model adapter
