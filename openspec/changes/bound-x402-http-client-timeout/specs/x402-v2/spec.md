## ADDED Requirements

### Requirement: Bounded X402 HTTP client timeout
The X402 interceptor SHALL create its wrapped HTTP client with a finite default timeout so automatic payment fetches cannot rely on an unbounded `http.Client`.

#### Scenario: X402 HTTP client has a default timeout
- **WHEN** `Interceptor.HTTPClient(ctx)` creates the wrapped payment client
- **THEN** the returned HTTP client SHALL have a non-zero timeout
- **AND** the timeout SHALL be at least 15 seconds

#### Scenario: Cached X402 HTTP client remains bounded
- **WHEN** `Interceptor.HTTPClient(ctx)` is called more than once
- **THEN** the cached client SHALL be reused
- **AND** the cached client SHALL retain the bounded timeout
