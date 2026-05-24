## MODIFIED Requirements

### Requirement: Retry / dead-letter handling page describes the first bounded retry slice
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe that retry/dead-letter evidence persistence stays best-effort while evidence-write failures are raised as operational errors.

#### Scenario: Retry docs mention elevated evidence-write failures
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find receipt-evidence write failures described as operational errors even when the retry hook remains best-effort

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the landed CLI pagination, explicit actor override, and JSON error payload behavior.

#### Scenario: Dead-letter browsing docs mention automation flags
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find dead-letter CLI `offset` / `limit` pagination described
- **AND** they SHALL find dead-letter CLI retry described as supporting an explicit `--actor` override
- **AND** they SHALL find machine-mode dead-letter CLI failures described as structured JSON error payloads when `--output json` is selected
