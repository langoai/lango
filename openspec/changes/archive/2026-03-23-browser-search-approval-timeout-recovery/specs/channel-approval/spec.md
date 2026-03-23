## MODIFIED Requirements

### Requirement: Turn-local approval replay protection
Each request SHALL maintain turn-local approval state keyed by `tool name + canonical params JSON`. The approval middleware SHALL consult this state before issuing a new approval request. Canonical params MAY normalize or omit fields that do not change approval risk for a tool.

#### Scenario: Turn-local positive replay
- **WHEN** a request already approved a specific `tool + params` once in the current turn
- **THEN** an identical retry in the same turn SHALL execute without issuing another approval prompt

#### Scenario: Canonical browser search replay key ignores limit-only variants
- **WHEN** `browser_search` is retried with the same query but different `limit` values or whitespace-only query differences
- **THEN** the approval middleware SHALL treat those retries as the same canonical approval action

#### Scenario: Turn-local denied or unavailable replay block
- **WHEN** a request already received deny or unavailable for a specific canonical approval action in the current turn
- **THEN** an identical retry in the same turn SHALL return the same failure immediately without issuing another approval prompt

#### Scenario: Timeout allows bounded re-prompt
- **WHEN** a request already timed out for a specific canonical approval action in the current turn
- **AND** the timeout count for that canonical action is below the configured per-turn timeout budget
- **THEN** the middleware SHALL issue another approval prompt instead of replay-blocking immediately

#### Scenario: Timeout replay blocked after budget exhaustion
- **WHEN** a request already accumulated the maximum per-turn timeout count for a specific canonical approval action
- **THEN** a later identical retry in the same turn SHALL return timeout immediately without issuing another approval prompt

#### Scenario: Different params require new approval
- **WHEN** the retried tool call changes the canonical approval action
- **THEN** the middleware SHALL treat it as a new approval request
