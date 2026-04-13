# Approval Policy (Delta)

## Changes

### Startup Warning

- **GIVEN** `security.interceptor.approvalPolicy` is set to `"none"`
- **WHEN** the application initializes
- **THEN** a WARN-level log message SHALL be emitted indicating that all tool calls will execute without user confirmation
- **AND** the message SHALL note this is not recommended for production
