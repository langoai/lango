## ADDED Requirements

### Requirement: --mode flag propagation to plain chat
The `lango chat` command SHALL read the inherited `--mode` persistent root flag and pass it to `runChat`. When a valid mode name is provided, the session SHALL be pre-created with that mode set.

#### Scenario: Launch plain chat with --mode
- **WHEN** user runs `lango chat --mode code-review`
- **THEN** the session SHALL be created with mode `code-review` set
- **AND** the first turn SHALL use the `code-review` mode's tool catalog and system hint

#### Scenario: Invalid mode name
- **WHEN** user runs `lango chat --mode nonexistent`
- **THEN** the command SHALL return an error message listing valid modes

### Requirement: Plain chat token usage for /cost
The plain chat path SHALL subscribe to `TokenUsageEvent` via EventBus and accumulate per-turn token counts. On turn completion (`DoneMsg`), accumulated tokens SHALL be emitted as `TurnTokenUsageMsg` so the `/cost` slash command reports accurate values.

#### Scenario: /cost after multiple turns in plain chat
- **WHEN** user runs 3 turns in `lango chat` then types `/cost`
- **THEN** the output SHALL show cumulative input tokens, output tokens, and estimated cost
- **AND** the values SHALL be non-zero

#### Scenario: Turn error resets accumulator
- **WHEN** a turn fails with `ErrorMsg`
- **THEN** the per-turn token accumulator SHALL be reset
- **AND** `turnActive` SHALL be set to false

### Requirement: Retry guard allows stale-stream recovery
The turn runner's retry loop SHALL allow retry after a stale-stream timeout even if chunks were already emitted. Only genuine mid-stream crashes (chunks emitted without stale timeout) SHALL block retry.

#### Scenario: Stale stream after partial output
- **WHEN** a provider emits some chunks then stops responding for longer than `staleTimeout`
- **THEN** the stale timer SHALL fire and cancel the attempt
- **AND** the retry loop SHALL retry (staleTriggered=true overrides chunksEmitted guard)

#### Scenario: Mid-stream crash without stale
- **WHEN** a provider emits some chunks then returns a retryable error immediately
- **THEN** the retry loop SHALL NOT retry (chunksEmitted=true, staleTriggered=false)
- **AND** the partial output SHALL remain visible to the user
