## MODIFIED Requirements

### Requirement: DataRoot enforces data path boundaries
The Config SHALL include a `DataRoot` field (default: `~/.lango/`) that defines the root directory for all lango data files. All configurable data paths (session.databasePath, graph.databasePath, skill.skillsDir, workflow.stateDir, p2p.keyDir, p2p.zkp.proofCacheDir, p2p.workspace.dataDir) MUST reside under DataRoot. The `NormalizePaths()` function SHALL expand tildes and resolve relative paths under DataRoot. The `ValidateDataPaths()` function SHALL reject any path outside DataRoot with a clear error message.

#### Scenario: Default paths pass validation
- **WHEN** config uses default paths (all under ~/.lango/)
- **THEN** NormalizePaths and ValidateDataPaths succeed

#### Scenario: External path rejected
- **WHEN** graph.databasePath is set to "/tmp/graph.db"
- **THEN** ValidateDataPaths returns error "graph.databasePath must be under data root"

#### Scenario: Relative path resolved under DataRoot
- **WHEN** graph.databasePath is set to "graph.db" (relative)
- **THEN** NormalizePaths resolves it to `<DataRoot>/graph.db`

#### Scenario: Custom DataRoot accepted
- **WHEN** DataRoot is set to "/data/lango" and all paths are under it
- **THEN** validation passes

### Requirement: ExecToolConfig supports additional protected paths
The ExecToolConfig SHALL include an `AdditionalProtectedPaths` field that specifies extra paths for CommandGuard to protect, in addition to DataRoot.

#### Scenario: Additional path protected
- **WHEN** additionalProtectedPaths includes "/var/secrets"
- **THEN** exec commands accessing /var/secrets are blocked
