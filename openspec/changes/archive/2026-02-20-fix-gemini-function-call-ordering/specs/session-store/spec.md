## MODIFIED Requirements

### Requirement: ToolCall stores FunctionResponse output
The `ToolCall` struct's `Output` field SHALL be used to store serialized `FunctionResponse.Response` data for tool/function role messages. This enables round-trip preservation of FunctionResponse metadata through the save/restore cycle without database schema changes.

#### Scenario: Tool message ToolCall with Output
- **WHEN** a tool message is stored with `ToolCalls` containing `Output` data
- **AND** the message is later loaded from the database
- **THEN** the `ToolCall.Output` field SHALL contain the original serialized response JSON
- **AND** `ToolCall.ID` and `ToolCall.Name` SHALL match the original FunctionResponse metadata
