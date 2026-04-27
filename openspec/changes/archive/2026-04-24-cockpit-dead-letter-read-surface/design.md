## Design Summary

This slice adds a read-only cockpit master-detail surface for post-adjudication dead letters.

The cockpit page has two panes:

- backlog table
- selected transaction detail pane

Data sources:

- backlog table reuses `list_dead_lettered_post_adjudication_executions`
- detail pane reuses `get_post_adjudication_execution_status`

Interaction remains intentionally narrow:

- selection only
- no filters
- no replay/write controls

The goal is to turn the landed dead-letter read model into a usable operator surface without introducing new backend plumbing.
