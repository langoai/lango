## MODIFIED Requirements

### Requirement: Learning suggestions render as actionable proposed missions
Mission Control SHALL back proposed missions with a transient proposal registry instead of rendering raw learning-buffer rows directly. In this Wave 3 slice, `LearningSuggestionEvent` is the only active proposal producer. Proposed missions SHALL remain transient and SHALL move through explicit proposal states such as `suggested`, `preparing`, and `prepared` before acceptance or dismissal.

#### Scenario: Buffered learning suggestion text is replay-safe
- **WHEN** learning suggestion patterns, proposed rules, or rationales contain ANSI/OSC escape sequences or embedded newlines before buffering
- **THEN** the cockpit learning suggestion buffer SHALL strip those control sequences
- **AND** it SHALL normalize the stored suggestion text to a single line before replay
