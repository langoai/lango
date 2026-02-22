## MODIFIED Requirements

### Requirement: Embedded prompt file content
The embedded AGENTS.md SHALL reference "ten tool categories" (previously nine) and include a P2P Network entry in the tool category list. The TOOL_USAGE.md SHALL include a P2P Networking Tool section after the existing Error Handling section.

#### Scenario: Tool category count updated
- **WHEN** the AGENTS.md embedded content is loaded
- **THEN** it contains the text "ten tool categories"

#### Scenario: P2P tool usage section present
- **WHEN** the TOOL_USAGE.md embedded content is loaded
- **THEN** it contains a "### P2P Networking Tool" section

### Requirement: Prompt test compatibility
The defaults_test.go SHALL assert "ten tool categories" instead of "nine tool categories" to match the updated embedded content.

#### Scenario: Test passes with updated count
- **WHEN** `go test ./internal/prompt/...` is run
- **THEN** all tests pass including the embedded content verification
