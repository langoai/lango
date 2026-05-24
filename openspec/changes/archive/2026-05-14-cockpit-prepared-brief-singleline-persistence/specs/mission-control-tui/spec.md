## MODIFIED Requirements

### Requirement: Learning suggestions render as actionable proposed missions
Mission Control SHALL back proposed missions with a transient proposal registry instead of rendering raw learning-buffer rows directly. In this Slice 3 slice, `LearningSuggestionEvent` is the only active proposal producer. Proposed missions SHALL remain transient and SHALL move through explicit proposal states such as `suggested`, `preparing`, and `prepared` before acceptance or dismissal.

#### Scenario: Prepared brief persistence stays single-line
- **WHEN** a prepared brief contributes multiple description segments during durable mission acceptance
- **THEN** Mission Control SHALL collapse those segments into one single-line persisted description
