## MODIFIED Requirements

### Requirement: Automatic post-adjudication execution page describes the first inline orchestration slice
The `docs/architecture/automatic-post-adjudication-execution.md` page SHALL describe the shared execution-mode runtime around inline post-adjudication execution.

#### Scenario: Automatic execution page describes the runtime default
- **WHEN** a user reads `docs/architecture/automatic-post-adjudication-execution.md`
- **THEN** they SHALL find that `auto_execute=true` selects inline execution
- **AND** they SHALL find that omitted execution flags default to `manual_recovery`
- **AND** they SHALL find that `auto_execute` and `background_execute` are mutually exclusive

### Requirement: Background post-adjudication execution page describes the shared execution-mode policy
The `docs/architecture/background-post-adjudication-execution.md` page SHALL describe background execution as one branch of the shared post-adjudication execution policy.

#### Scenario: Background page describes manual/background/inline alignment
- **WHEN** a user reads `docs/architecture/background-post-adjudication-execution.md`
- **THEN** they SHALL find the shared `manual_recovery` / `inline` / `background` modes described
- **AND** they SHALL find that `background_execute=true` selects background execution while omitted flags still default to `manual_recovery`

### Requirement: Retry / dead-letter handling page describes the normalized retry policy shape
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe the normalized runtime retry policy used by background post-adjudication recovery.

#### Scenario: Retry page describes bounded retry policy fields
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find the retry policy fields `MaxRetryAttempts` and `BaseDelay` described
- **AND** they SHALL find bounded retry scheduling with exponential backoff described
- **AND** they SHALL find the shared `post_adjudication_retry` evidence source with `retry-scheduled` and `dead-lettered` subtypes described

### Requirement: Operator replay / manual retry page describes replay as part of the recovery substrate
The `docs/architecture/operator-replay-manual-retry.md` page SHALL describe operator replay as part of the same recovery substrate used by automatic retry and dead-letter handling.

#### Scenario: Replay page describes shared recovery evidence
- **WHEN** a user reads `docs/architecture/operator-replay-manual-retry.md`
- **THEN** they SHALL find that replay requires dead-letter evidence from the shared recovery source
- **AND** they SHALL find `manual-retry-requested` evidence described
- **AND** they SHALL find that replay reuses the background dispatch path without clearing prior dead-letter evidence

### Requirement: Policy-driven replay controls page describes authorization on top of the shared recovery gate
The `docs/architecture/policy-driven-replay-controls.md` page SHALL describe replay authorization as a config-backed allowlist layer on top of the shared recovery evidence gate.

#### Scenario: Replay policy page describes shared gate semantics
- **WHEN** a user reads `docs/architecture/policy-driven-replay-controls.md`
- **THEN** they SHALL find `replay.allowed_actors`, `replay.release_allowed_actors`, and `replay.refund_allowed_actors` described
- **AND** they SHALL find fail-closed actor resolution described
- **AND** they SHALL find that authorization sits on top of the shared recovery gate described

### Requirement: P2P knowledge exchange track reflects the landed replay / recovery runtime alignment
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the replay / recovery runtime alignment as landed work and narrow the remaining work accordingly.

#### Scenario: Track page removes already-landed runtime gaps from remaining work
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the shared post-adjudication execution-mode policy described as landed work
- **AND** they SHALL find the normalized retry / dead-letter policy shape described as landed work
- **AND** they SHALL find replay described as using the same recovery substrate as automatic retry and dead-letter handling
- **AND** they SHALL find the remaining work narrowed to policy editing, broader retry/recovery substrate reuse, and broader dispute integration
