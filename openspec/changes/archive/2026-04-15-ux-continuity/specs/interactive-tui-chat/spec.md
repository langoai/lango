## ADDED Requirements

### Requirement: Compaction status entry in transcript
The TUI SHALL subscribe to `CompactionCompletedEvent` and `CompactionSlowEvent` and render a transient status entry in the chat transcript.

- `CompactionCompletedEvent` SHALL render as an `itemStatus` entry with a concise message such as `"context compacted (reclaimed N tokens)"`.
- `CompactionSlowEvent` SHALL render as an `itemStatus` entry with a warn-styled message such as `"compaction still running — proceeded with current context"`.

These entries SHALL NOT block streaming and SHALL NOT be persisted as assistant or user messages.

#### Scenario: Completed event appended as status
- **WHEN** a `CompactionCompletedEvent` with `ReclaimedTokens=4200` is received
- **THEN** an `itemStatus` entry SHALL be appended with text indicating the reclaimed token count
- **AND** the entry SHALL NOT appear in the session's persisted message history

#### Scenario: Slow event appended as warn status
- **WHEN** a `CompactionSlowEvent` is received
- **THEN** an `itemStatus` entry styled as a warning SHALL be appended
- **AND** the chat viewport SHALL remain responsive

### Requirement: Learning suggestion rendering in TUI
The TUI SHALL subscribe to `LearningSuggestionEvent` and render the suggestion as an inline approval prompt reusing the existing approval rendering path. Approval/denial SHALL route through the existing approval pipeline, producing the same persistence outcome whether the approval is resolved via TUI or a channel surface.

#### Scenario: Suggestion renders as approval prompt
- **WHEN** a `LearningSuggestionEvent` with confidence 0.6 is published while on the chat page
- **THEN** the TUI SHALL render an approval prompt summarizing the proposed rule and confidence
- **AND** the user SHALL be able to accept or deny via the same keys used for tool approvals (`a`/`d`/`s`)

#### Scenario: Acceptance persists learning via approval pipeline
- **WHEN** the user accepts a learning suggestion prompt
- **THEN** the approval pipeline SHALL route acceptance to the learning engine's persistence path
- **AND** the stored confidence SHALL equal the suggestion's confidence value (not auto-boosted)

#### Scenario: Denial suppresses re-emission within dedup window
- **WHEN** the user denies a learning suggestion prompt
- **THEN** the pattern hash SHALL be recorded as "dismissed" for the configured dedup window
- **AND** no new prompt for the same pattern SHALL appear within that window
