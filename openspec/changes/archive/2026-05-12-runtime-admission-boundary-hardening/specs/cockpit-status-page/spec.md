## ADDED Requirements

### Requirement: Graph admission metrics are surfaced on the cockpit status page
The cockpit status page SHALL surface observe-only graph admission metrics from the runtime feedback snapshot, including graph-admission counts grouped by source and validator identity, extractor dropped-unknown baselines, unmapped-source counts, and aggregate graph write-failure baselines.

#### Scenario: Status page renders graph admission metrics
- **WHEN** the cockpit status page is rendered while observe mode metrics are available
- **THEN** it SHALL display event-bus graph-admission counts grouped by supported producer source and producer group, plus a separate grouped view for the synthetic `content_saved_extractor` source label
- **AND** it SHALL display validator-source as a grouping key on graph-admission metrics rather than as a separate independent metric family
- **AND** it SHALL display extractor dropped-unknown, unmapped-source, and aggregate graph write-failure baseline counts as distinct metrics
- **AND** it SHALL preserve raw `unmapped-source` identity by grouping those counts by raw source label
- **AND** it SHALL preserve validator-source identity by grouping those counts by validator-source identifier
- **AND** it SHALL display `known`, `unknown`, and `unvalidated` triple totals for graph-admission decisions
