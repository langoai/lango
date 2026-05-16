## MODIFIED Requirements

### Requirement: Learning suggestions render as actionable proposed missions
Mission Control SHALL back proposed missions with a transient proposal registry instead of rendering raw learning-buffer rows directly. In this Wave 3 slice, `LearningSuggestionEvent` is the only active proposal producer. Proposed missions SHALL remain transient and SHALL move through explicit proposal states such as `suggested`, `preparing`, and `prepared` before acceptance or dismissal.

#### Scenario: Accepted proposal description stays plain and single-line
- **WHEN** prepared-brief or fallback proposal summary text contains ANSI/OSC escape sequences or embedded newlines before durable mission acceptance
- **THEN** Mission Control SHALL strip those control sequences
- **AND** it SHALL normalize the persisted mission description text to a single line before calling mission acceptance
