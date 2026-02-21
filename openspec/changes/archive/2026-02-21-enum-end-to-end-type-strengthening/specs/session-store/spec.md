## MODIFIED Requirements

### Requirement: Message structure
The `session.Message` struct SHALL use `types.MessageRole` for its `Role` field instead of plain `string`. All internal code that reads or writes `Message.Role` SHALL use typed enum constants (`types.RoleUser`, `types.RoleAssistant`, `types.RoleTool`, `types.RoleFunction`, `types.RoleModel`). The `string()` cast SHALL only occur at system boundaries: Ent DB writes (`SetRole(string(msg.Role))`), Ent DB reads (`types.MessageRole(m.Role)`), and external API mapping (genai `Content.Role`).

#### Scenario: Message role uses typed enum
- **WHEN** a `session.Message` is created anywhere in internal code
- **THEN** the `Role` field SHALL be assigned a `types.MessageRole` constant, not a raw string literal

#### Scenario: DB boundary cast on write
- **WHEN** a message is persisted to the Ent store via `SetRole()`
- **THEN** the role SHALL be cast to `string` at the call site: `SetRole(string(msg.Role))`

#### Scenario: DB boundary cast on read
- **WHEN** a message is loaded from the Ent store
- **THEN** the role SHALL be cast from `string` to `types.MessageRole`: `Role: types.MessageRole(m.Role)`

#### Scenario: JSON serialization backward compatibility
- **WHEN** a `session.Message` with `Role: types.RoleUser` is serialized to JSON
- **THEN** the JSON output SHALL contain `"role":"user"` (unchanged from previous format)
