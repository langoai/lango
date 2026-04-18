## ADDED Requirements

### Requirement: Runtime readers use broker-backed storage capabilities
Runtime app and CLI reader paths MUST be able to obtain learning history, pending inquiries, workflow state, alert history, and reputation details through broker-backed storage capabilities.

#### Scenario: Reader path resolved through broker capability
- **WHEN** a runtime app or CLI path needs one of those reader surfaces while broker mode is active
- **THEN** it can obtain the data through broker-backed storage capabilities without opening or querying the application DB directly in the parent process
