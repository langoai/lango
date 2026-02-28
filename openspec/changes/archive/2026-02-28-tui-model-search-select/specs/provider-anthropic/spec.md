## MODIFIED Requirements

### Requirement: Live model listing
The Anthropic provider's `ListModels()` MUST call the Anthropic Models API instead of returning hardcoded values.

#### Scenario: Successful model listing
- **WHEN** ListModels is called with valid API credentials
- **THEN** returns all models from the API using paginated auto-paging with limit 1000

#### Scenario: Partial failure
- **WHEN** API returns some models before encountering an error
- **THEN** returns the successfully fetched models without error

#### Scenario: Complete failure
- **WHEN** API call fails with no models retrieved
- **THEN** returns error with wrapped context
