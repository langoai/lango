## ADDED Requirements

### Requirement: Telegram download request shape is verified

Telegram media download coverage SHALL verify not only success and failure outcomes, but also the outgoing HTTP request contract.

#### Scenario: Telegram download uses HTTP GET with a timeout-backed context
- **WHEN** the Telegram file download regression exercises `DownloadFile`
- **THEN** the outgoing request SHALL use the HTTP GET method
- **AND** the request context SHALL carry a deadline derived from the 30-second download timeout contract
