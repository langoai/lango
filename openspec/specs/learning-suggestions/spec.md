# learning-suggestions Specification

## Purpose
TBD - created by archiving change ux-continuity. Update Purpose after archive.
## Requirements
### Requirement: Suggestion emission threshold
The `learning.SuggestionEngine` SHALL emit a `LearningSuggestionEvent` whenever a learning candidate crosses the configured suggestion confidence threshold (default `0.5`) AND does not already match an existing auto-applied learning (confidence ≥ 0.7). The suggestion threshold SHALL be distinct from the auto-apply threshold and SHALL be configurable under `learning.suggestions.threshold` (valid range `[0.1, 0.9]`).

#### Scenario: Above threshold emits suggestion
- **WHEN** a new learning candidate reaches confidence 0.55 and no existing learning with confidence ≥ 0.7 matches the same pattern
- **AND** `learning.suggestions.enabled` is `true`
- **THEN** a `LearningSuggestionEvent` SHALL be published

#### Scenario: Below threshold does not emit
- **WHEN** a learning candidate has confidence 0.4
- **THEN** no suggestion SHALL be published

#### Scenario: Already auto-applied pattern does not emit
- **WHEN** an existing learning with confidence 0.8 matches the same pattern
- **THEN** no suggestion SHALL be published

#### Scenario: Feature disabled suppresses emission
- **WHEN** `learning.suggestions.enabled` is `false`
- **THEN** no suggestion SHALL be published

### Requirement: Suggestion rate limit and dedup
Suggestion emission SHALL be rate-limited (default 1 suggestion per 10 turns per session) and SHALL deduplicate by pattern hash within a configurable window (default `1 hour`). Both SHALL be configurable under `learning.suggestions.rateLimit` and `learning.suggestions.dedupWindow`.

#### Scenario: Rate limit suppresses burst
- **WHEN** three candidates qualify within the same turn
- **AND** the rate limit is 1 per 10 turns
- **THEN** only the highest-confidence one SHALL be emitted
- **AND** the others SHALL be dropped with a debug-level log

#### Scenario: Dedup within window
- **WHEN** a suggestion for pattern `P` was emitted 10 minutes ago
- **AND** the dedup window is 1 hour
- **AND** the same pattern qualifies again
- **THEN** no new suggestion SHALL be emitted

### Requirement: LearningSuggestionEvent
The eventbus SHALL carry a `LearningSuggestionEvent` with fields: `SessionKey string`, `SuggestionID string`, `Pattern string`, `ProposedRule string`, `Confidence float64`, `Rationale string`, and `Timestamp time.Time`. The event's `EventName()` SHALL return `"learning.suggestion"`.

#### Scenario: Event contract
- **WHEN** a `LearningSuggestionEvent` is created with confidence 0.55
- **THEN** `EventName()` SHALL return `"learning.suggestion"`
- **AND** the event SHALL carry the session key, stable suggestion ID, pattern, proposed rule, confidence, rationale, and timestamp

### Requirement: Approval-gated persistence
Accepting a learning suggestion SHALL go through the existing approval pipeline. On approval, the suggestion SHALL be persisted via the learning engine's normal save path with confidence set to the suggestion's confidence (not auto-boosted to auto-apply territory). On denial, the suggestion SHALL be recorded as a "dismissed" entry keyed by pattern hash to suppress immediate re-emission within the dedup window.

#### Scenario: User approves suggestion
- **WHEN** the user approves a `LearningSuggestionEvent` via TUI or channel approval surface
- **THEN** the learning engine SHALL persist the rule using its normal save path
- **AND** the stored confidence SHALL equal the suggestion's confidence value

#### Scenario: User denies suggestion
- **WHEN** the user denies a `LearningSuggestionEvent`
- **THEN** the suggestion SHALL NOT be persisted as a learning
- **AND** the pattern hash SHALL be recorded as "dismissed" so future matches within the dedup window are suppressed

#### Scenario: Approval disabled suppresses surface
- **WHEN** approval infrastructure is disabled in config
- **THEN** suggestion subscribers SHALL NOT render a UI prompt
- **AND** the event SHALL still be published (debug-only consumers may record it)

### Requirement: Multi-channel suggestion surface
Suggestion rendering SHALL be implemented through the eventbus such that at least two distinct subscriber paths exist: the TUI chat surface and the channel adapter layer. Neither path SHALL short-circuit the other, and no suggestion rendering SHALL live exclusively inside the TUI package.

#### Scenario: TUI renders suggestion as chat status
- **WHEN** a `LearningSuggestionEvent` is published while the user is on the TUI chat page
- **THEN** the TUI SHALL render a status entry summarizing the suggestion
- **AND** SHALL produce an approval prompt using the existing approval rendering path

#### Scenario: Channel adapter renders suggestion in-channel
- **WHEN** a `LearningSuggestionEvent` is published for a session that originated from a channel
- **THEN** the channel adapter SHALL deliver a channel-native message (e.g., Slack block, Telegram message) with approve/deny actions
- **AND** user response SHALL route to the same approval pipeline as any other approval request

### Requirement: Config surface for learning suggestions
The system SHALL provide additive fields under `learning.suggestions`: `enabled bool` (default `true`), `threshold float64` (default `0.5`, valid range `[0.1, 0.9]`), `rateLimit int` (turns between suggestions per session, default `10`, valid range `[1, 100]`), and `dedupWindow time.Duration` (default `1h`, valid range `[1m, 24h]`). Invalid values SHALL be clamped to valid ranges with a warning log.

#### Scenario: Defaults when unset
- **WHEN** no `learning.suggestions.*` config is set
- **THEN** suggestions SHALL be enabled with threshold 0.5, rate limit 10, and dedup window 1 hour

