## MODIFIED Requirements

### REQ-EMB-001: Embedding Provider Interface
The system SHALL provide an `EmbeddingProvider` interface supporting batch text-to-vector conversion with provider ID, embed, and dimensions methods. Each provider SHALL pass the configured `dimensions` value to its underlying API call so that returned vectors match the configured dimension.

#### Scenario: Google provider returns vectors matching configured dimensions
- **WHEN** GoogleProvider is created with dimensions=128 and Embed is called
- **THEN** the `EmbedContent` API call SHALL include `OutputDimensionality` set to 128 and returned vectors SHALL have exactly 128 dimensions

#### Scenario: OpenAI provider returns vectors matching configured dimensions
- **WHEN** OpenAIProvider is created with dimensions=128 and Embed is called
- **THEN** the `EmbeddingRequest` SHALL include `Dimensions` set to 128 and returned vectors SHALL have exactly 128 dimensions

#### Scenario: Local provider returns vectors matching configured dimensions
- **WHEN** LocalProvider is created with dimensions=128 and Embed is called
- **THEN** the `EmbeddingRequest` SHALL include `Dimensions` set to 128 and returned vectors SHALL have exactly 128 dimensions

#### Scenario: Vectors match SQLite vec table schema
- **WHEN** any provider returns embedding vectors
- **THEN** the vector dimension SHALL match the SQLite vec table's configured `float[N]` dimension to prevent insert and search failures
