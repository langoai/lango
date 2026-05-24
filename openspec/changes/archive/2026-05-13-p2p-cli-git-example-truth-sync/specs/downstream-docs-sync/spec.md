## ADDED Requirements

### Requirement: P2P CLI git-bundle examples stay truth-aligned

Public P2P CLI docs SHALL show workspace/git runtime examples using the actual registered tool names and parameter names.

#### Scenario: P2P git example uses actual workspace tool names
- **WHEN** a user reads the git-bundle workflow example in the P2P CLI docs
- **THEN** the example SHALL use `workspaceId` instead of `workspace_id`
- **AND** SHALL reference only the currently registered runtime tools such as `p2p_git_init`, `p2p_git_push`, `p2p_git_log`, and `p2p_git_diff`
