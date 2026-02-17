## ADDED Requirements

### Requirement: Context Layer Architecture
The system SHALL organize context into distinct layers for retrieval-augmented generation, including a Pending Inquiries layer for the proactive librarian.

#### Scenario: Context layer definitions
- **WHEN** the context retriever is initialized
- **THEN** the system SHALL recognize 9 layers: Tool Registry, User Knowledge, Skill Patterns, External Knowledge, Agent Learnings, Runtime Context, Observations, Reflections, and Pending Inquiries

#### Scenario: Pending inquiries layer retrieval
- **WHEN** Retrieve is called with LayerPendingInquiries in the requested layers
- **THEN** the retriever delegates to the InquiryProvider to fetch pending inquiry items

#### Scenario: No inquiry provider configured
- **WHEN** LayerPendingInquiries is requested but no InquiryProvider is set
- **THEN** the retriever returns nil items for that layer without error

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
  - "Runtime Context" for runtime context items
  - "Available Tools" for tool registry items
  - "User Knowledge" for user knowledge items
  - "Known Solutions" for agent learnings
  - "Available Skills" for skill patterns
  - "External References" for external knowledge

#### Scenario: Section ordering
- **WHEN** `AssemblePrompt` is called with items from multiple layers
- **THEN** "Runtime Context" SHALL appear before "Available Tools"
- **AND** "Available Tools" SHALL appear before "User Knowledge"

#### Scenario: Observation memory section
- **WHEN** `AssemblePrompt` is called with observations or reflections
- **THEN** the system SHALL append a "Conversation Memory" section after knowledge sections
- **AND** reflections SHALL appear before observations within that section

#### Scenario: Inquiries present in context
- **WHEN** AssemblePrompt is called with LayerPendingInquiries items
- **THEN** the output includes a "## Pending Knowledge Inquiries" section with each inquiry formatted as `- [topic] question (context: why)`

#### Scenario: No inquiries
- **WHEN** no LayerPendingInquiries items exist
- **THEN** no inquiry section is included in the prompt

### Requirement: Context-Aware Model Adapter
The system SHALL wrap the ADK model adapter to transparently inject retrieved context and observational memory.

#### Scenario: System prompt augmentation
- **WHEN** `GenerateContent` is called on the context-aware adapter
- **THEN** the system SHALL extract the user's latest message as query
- **AND** retrieve relevant context from all 6 layers (Runtime Context, Tool Registry, User Knowledge, Skill Patterns, External Knowledge, Agent Learnings)
- **AND** update the runtime adapter's session state before retrieval
- **AND** retrieve observations and reflections for the current session
- **AND** augment the system instruction with assembled context including observations
- **AND** forward the modified request to the underlying model adapter

### Requirement: Tool Registry Provider Interface
The system SHALL define a `ToolRegistryProvider` interface for supplying available tool information to the context retriever.

#### Scenario: List all tools
- **WHEN** `ListTools()` is called on a `ToolRegistryProvider`
- **THEN** the system SHALL return all registered `ToolDescriptor` entries

#### Scenario: Search tools by query
- **WHEN** `SearchTools(query, limit)` is called on a `ToolRegistryProvider`
- **THEN** the system SHALL return tools whose name or description contains the query (case-insensitive)
- **AND** the result SHALL contain at most `limit` items

### Requirement: Runtime Context Provider Interface
The system SHALL define a `RuntimeContextProvider` interface for supplying session and system state.

#### Scenario: Get runtime context
- **WHEN** `GetRuntimeContext()` is called on a `RuntimeContextProvider`
- **THEN** the system SHALL return a `RuntimeContext` containing session key, channel type, active tool count, encryption enabled flag, knowledge enabled flag, and memory enabled flag

### Requirement: Tool Registry Retrieval
The system SHALL retrieve matching tools when the Tool Registry layer is requested.

#### Scenario: Retrieve tools by keyword
- **WHEN** `Retrieve` is called with `LayerToolRegistry` in the layer list
- **AND** a `ToolRegistryProvider` is configured
- **THEN** the system SHALL search tools using extracted keywords and return matching items as `ContextItem` entries with `LayerToolRegistry` layer

#### Scenario: Nil tool provider
- **WHEN** `Retrieve` is called with `LayerToolRegistry` in the layer list
- **AND** no `ToolRegistryProvider` is configured
- **THEN** the system SHALL return zero items for that layer without error

### Requirement: Runtime Context Retrieval
The system SHALL retrieve session state when the Runtime Context layer is requested.

#### Scenario: Retrieve runtime context
- **WHEN** `Retrieve` is called with `LayerRuntimeContext` in the layer list
- **AND** a `RuntimeContextProvider` is configured
- **THEN** the system SHALL return a single `ContextItem` with key "session-state" containing formatted session information

#### Scenario: Nil runtime provider
- **WHEN** `Retrieve` is called with `LayerRuntimeContext` in the layer list
- **AND** no `RuntimeContextProvider` is configured
- **THEN** the system SHALL return zero items for that layer without error

### Requirement: Tool Registry Adapter
The system SHALL provide a `ToolRegistryAdapter` that adapts `[]*agent.Tool` to `ToolRegistryProvider`.

#### Scenario: Boundary copy on construction
- **WHEN** a `ToolRegistryAdapter` is created with a tool slice
- **THEN** the adapter SHALL copy the slice so external mutations do not affect internal state

#### Scenario: Case-insensitive search
- **WHEN** `SearchTools` is called with a query
- **THEN** the adapter SHALL match tools using case-insensitive substring comparison on name and description

### Requirement: Runtime Context Adapter
The system SHALL provide a `RuntimeContextAdapter` with thread-safe session updates.

#### Scenario: Default channel type
- **WHEN** a `RuntimeContextAdapter` is created without calling `SetSession`
- **THEN** the channel type SHALL be "direct"

#### Scenario: Derive channel type from session key
- **WHEN** `SetSession` is called with a session key in the format "channel:id:subid"
- **THEN** the adapter SHALL extract the channel type from the prefix
- **AND** recognized prefixes SHALL be "telegram", "discord", "slack"
- **AND** unrecognized prefixes SHALL map to "direct"

#### Scenario: Thread-safe access
- **WHEN** `SetSession` and `GetRuntimeContext` are called concurrently
- **THEN** the adapter SHALL use mutex protection to prevent data races

### Requirement: InquiryProvider Interface
The system SHALL define an InquiryProvider interface with method `PendingInquiryItems(ctx, sessionKey, limit) ([]ContextItem, error)`. The ContextRetriever SHALL accept an optional InquiryProvider via `WithInquiryProvider()`.

#### Scenario: Wire inquiry provider
- **WHEN** WithInquiryProvider is called with a non-nil provider
- **THEN** the retriever uses it for LayerPendingInquiries retrieval
