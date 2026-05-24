## MODIFIED Requirements

### Requirement: CLI index quick references include required operands
The CLI index SHALL list quick-reference commands with required positional
arguments and required flags for provenance and P2P reputation commands.

#### Scenario: Config CLI docs explain dynamic key placeholders
- **WHEN** a user reads `docs/cli/config.md`
- **THEN** the config keys section SHALL show dynamic map-backed templates such as `providers.<name>.apiKey`, `mcp.servers.<name>.env.<key>`, and `mcp.servers.<name>.headers.<key>`
- **AND** it SHALL explain what `<name>` and `<key>` represent
