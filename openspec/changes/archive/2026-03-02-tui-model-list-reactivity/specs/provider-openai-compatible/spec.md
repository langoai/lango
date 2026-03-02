## MODIFIED Requirements

### Requirement: ListModels debug logging
The OpenAI provider's `ListModels()` method SHALL log debug messages for request start, success (with model count), and failure (with error).

#### Scenario: Successful model listing logged
- **WHEN** `ListModels()` succeeds and returns models
- **THEN** a debug log SHALL be emitted with provider ID and model count

#### Scenario: Failed model listing logged
- **WHEN** `ListModels()` fails with an error
- **THEN** a debug log SHALL be emitted with provider ID and error details
