## ADDED Requirements

### Requirement: Public workbench docs describe context-sensitive post-turn defaults

Public workbench documentation SHALL explain that the post-turn default `Enter` starter depends on whether a workspace context is available.

#### Scenario: Docs mention generic-vs-repo post-turn default
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that post-turn defaults stay structure-oriented outside repo context and stay next-change-oriented inside detected workspaces
