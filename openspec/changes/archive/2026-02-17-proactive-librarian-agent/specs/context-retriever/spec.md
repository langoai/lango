## MODIFIED Requirements

### Requirement: Context Layer Enumeration
The system SHALL define context layers including: LayerToolRegistry, LayerUserKnowledge, LayerSkillPatterns, LayerExternalKnowledge, LayerAgentLearnings, LayerRuntimeContext, LayerObservations, LayerReflections, and LayerPendingInquiries. The LayerPendingInquiries layer SHALL inject pending knowledge inquiries from the proactive librarian.

#### Scenario: Pending inquiries layer retrieval
- **WHEN** Retrieve is called with LayerPendingInquiries in the requested layers
- **THEN** the retriever delegates to the InquiryProvider to fetch pending inquiry items

#### Scenario: No inquiry provider configured
- **WHEN** LayerPendingInquiries is requested but no InquiryProvider is set
- **THEN** the retriever returns nil items for that layer without error

### Requirement: Prompt Assembly with Inquiries
The system SHALL include a "Pending Knowledge Inquiries" section in the assembled prompt when pending inquiry items are present. The section SHALL instruct the agent to weave ONE question naturally into its response.

#### Scenario: Inquiries present in context
- **WHEN** AssemblePrompt is called with LayerPendingInquiries items
- **THEN** the output includes a "## Pending Knowledge Inquiries" section with each inquiry formatted as `- [topic] question (context: why)`

#### Scenario: No inquiries
- **WHEN** no LayerPendingInquiries items exist
- **THEN** no inquiry section is included in the prompt

## ADDED Requirements

### Requirement: InquiryProvider Interface
The system SHALL define an InquiryProvider interface with method `PendingInquiryItems(ctx, sessionKey, limit) ([]ContextItem, error)`. The ContextRetriever SHALL accept an optional InquiryProvider via `WithInquiryProvider()`.

#### Scenario: Wire inquiry provider
- **WHEN** WithInquiryProvider is called with a non-nil provider
- **THEN** the retriever uses it for LayerPendingInquiries retrieval
