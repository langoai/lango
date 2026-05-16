## ADDED Requirements

### Requirement: Default confirmation wrapper supports deterministic stream seams
The shared `prompt.Confirm(...)` helper SHALL allow its default input and output streams to be replaced in tests without changing runtime confirmation behavior.

#### Scenario: Wrapper uses injected streams under test
- **WHEN** `prompt.Confirm(...)` is exercised in tests with injected default input and output streams
- **THEN** the wrapper SHALL read from the injected input stream
- **AND** it SHALL write the confirmation prompt to the injected output stream

#### Scenario: Wrapper preserves existing confirmation semantics
- **WHEN** `prompt.Confirm(...)` receives `y` or `yes`
- **THEN** it SHALL return approval
- **AND** non-affirmative input SHALL continue to return denial

#### Scenario: Wrapper preserves existing runtime defaults
- **WHEN** production code calls `prompt.Confirm(...)` without overriding any seams
- **THEN** the helper SHALL continue to use the process default stdin and stdout streams
