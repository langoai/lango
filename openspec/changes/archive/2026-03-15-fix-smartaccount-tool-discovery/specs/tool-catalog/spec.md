## MODIFIED Requirements

### Requirement: Dispatcher tools
The system SHALL provide `BuildDispatcher(catalog)` returning three tools: `builtin_list`, `builtin_invoke`, and `builtin_health`.

#### Scenario: builtin_list returns tool catalog
- **WHEN** `builtin_list` is invoked with no parameters
- **THEN** it SHALL return all categories (including disabled ones) and all tools with their schemas
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

### Requirement: Safety levels
`builtin_list` SHALL have SafetyLevelSafe. `builtin_invoke` SHALL have SafetyLevelDangerous. `builtin_health` SHALL have SafetyLevelSafe.

#### Scenario: Safety level assignment
- **WHEN** `BuildDispatcher()` creates the dispatcher tools
- **THEN** `builtin_list` safety level SHALL be Safe
- **AND** `builtin_invoke` safety level SHALL be Dangerous
- **AND** `builtin_health` safety level SHALL be Safe

## ADDED Requirements

### Requirement: Disabled category registration
The system SHALL register disabled categories in the tool catalog when a subsystem is not initialized, so that `builtin_list` and `builtin_health` can report their existence and required config keys.

#### Scenario: Smart account disabled registers disabled category
- **WHEN** `initSmartAccount()` returns nil (smart account disabled or payment missing)
- **THEN** a `smartaccount` category SHALL be registered with `Enabled: false`
- **AND** the category description SHALL include instructions for enabling

#### Scenario: Disabled category visible in builtin_list
- **WHEN** `builtin_list` is invoked
- **THEN** disabled categories SHALL appear in the `categories` list with `enabled: false`
