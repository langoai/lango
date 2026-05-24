## MODIFIED Requirements

### Requirement: Workbench activity retains assistant reply summaries after a turn completes

The standalone workbench SHALL retain a concise assistant reply summary in its activity lane after a turn completes so the operator can still see what happened from the workbench shell.

#### Scenario: Long or multi-line replies are compacted for activity
- **WHEN** a workbench turn completes with a long or multi-line assistant reply summary
- **THEN** the activity lane SHALL normalize that summary into one compact line
- **AND** SHALL keep it bounded as a short timeline summary rather than replaying the full raw reply
