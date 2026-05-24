## ADDED Requirements

### Requirement: Built-in inspection tools classify activity semantically
Built-in tools that expose scheduler, workspace, or repository inspection state SHALL publish read/query capability activity that matches the underlying operation instead of defaulting to mutation-oriented categories.

#### Scenario: Cron inspection tools are query-classified
- **WHEN** `cron_list` or `cron_history` is registered
- **THEN** each tool SHALL be marked `ReadOnly=true`
- **AND** each tool SHALL classify its primary capability activity as `query`

#### Scenario: Workspace inspection tools distinguish query from read
- **WHEN** `p2p_workspace_list` or `p2p_workspace_status` is registered
- **THEN** each tool SHALL be marked `ReadOnly=true`
- **AND** each tool SHALL classify its primary capability activity as `query`
- **WHEN** `p2p_workspace_read`, `p2p_git_log`, `p2p_git_diff`, or `p2p_git_leaves` is registered
- **THEN** each tool SHALL be marked `ReadOnly=true`
- **AND** each tool SHALL classify its primary capability activity as `read`
