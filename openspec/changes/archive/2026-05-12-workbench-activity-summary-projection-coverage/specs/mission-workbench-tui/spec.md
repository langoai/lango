## MODIFIED Requirements

### Requirement: Workbench activity retains assistant reply summaries after a turn completes

The standalone workbench SHALL retain a concise assistant reply summary in its activity lane after a turn completes so the operator can still see what happened from the workbench shell.

#### Scenario: Activity projection preserves compact reply summaries
- **WHEN** a compacted assistant activity summary is projected into the Mission Control activity view
- **THEN** the projected activity row SHALL preserve the one-line bounded summary instead of re-expanding the raw reply text
