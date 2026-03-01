## MODIFIED Requirements

### Requirement: Lango CLI block message includes builtin_list hint
The `blockLangoExec` catch-all message for unrecognized `lango` subcommands SHALL include a hint to use `builtin_list` for tool discovery.

#### Scenario: Catch-all message with builtin_list hint
- **WHEN** an unrecognized `lango` subcommand is blocked by `blockLangoExec`
- **THEN** the returned message SHALL contain "builtin_list"
- **AND** SHALL suggest using built-in tools or asking the user to run the command directly
