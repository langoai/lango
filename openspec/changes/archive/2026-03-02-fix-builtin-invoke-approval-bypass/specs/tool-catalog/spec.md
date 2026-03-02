## MODIFIED Requirements

### Requirement: Dispatcher tools
The system SHALL provide `BuildDispatcher(catalog)` returning two tools: `builtin_list` and `builtin_invoke`.

#### Scenario: builtin_list returns tool catalog
- **WHEN** `builtin_list` is invoked with no parameters
- **THEN** it SHALL return all categories and all tools with their schemas
- **AND** the total count of registered tools

#### Scenario: builtin_list filters by category
- **WHEN** `builtin_list` is invoked with `category: "exec"`
- **THEN** it SHALL return only tools in the "exec" category

#### Scenario: builtin_invoke executes a safe registered tool
- **WHEN** `builtin_invoke` is invoked with a tool_name whose SafetyLevel is less than Dangerous
- **THEN** it SHALL execute the tool's handler and return `{tool, result}`

#### Scenario: builtin_invoke blocks dangerous tools
- **WHEN** `builtin_invoke` is invoked with a tool_name whose SafetyLevel is Dangerous or higher
- **THEN** it SHALL return an error containing "requires approval" and "delegate to the appropriate sub-agent"
- **AND** it SHALL NOT execute the tool's handler

#### Scenario: builtin_invoke rejects unknown tool
- **WHEN** `builtin_invoke` is invoked with a tool_name not in the catalog
- **THEN** it SHALL return an error containing "not found in catalog"
