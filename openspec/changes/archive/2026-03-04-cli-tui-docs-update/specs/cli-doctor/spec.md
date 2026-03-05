## MODIFIED Requirements

### Requirement: Tool hooks health check
The doctor command SHALL include a ToolHooksCheck that validates hook system configuration. The check SHALL implement the Name()/Run()/Fix() interface. The check SHALL skip when hooks.enabled is false. When enabled, it SHALL verify that at least one hook type is active. It SHALL warn when securityFilter is enabled but blockedCommands is empty.

#### Scenario: Hooks disabled
- **WHEN** doctor runs with hooks.enabled set to false
- **THEN** ToolHooksCheck returns StatusSkip with message "Hooks are disabled"

#### Scenario: Hooks enabled with no active hooks
- **WHEN** hooks.enabled is true but all hook types (securityFilter, accessControl, eventPublishing, knowledgeSave) are false
- **THEN** ToolHooksCheck returns StatusWarn with message indicating no hook types are active

#### Scenario: Security filter without blocked commands
- **WHEN** hooks.enabled is true and securityFilter is true but blockedCommands is empty
- **THEN** ToolHooksCheck returns StatusWarn indicating security filter has no blocked commands configured

### Requirement: Agent registry health check
The doctor command SHALL include an AgentRegistryCheck that validates multi-agent registry configuration. The check SHALL skip when agent.multiAgent is false. When enabled, it SHALL verify that at least one sub-agent type is configured and that agent.provider is set.

#### Scenario: Multi-agent disabled
- **WHEN** doctor runs with agent.multiAgent set to false
- **THEN** AgentRegistryCheck returns StatusSkip with message "Multi-agent is disabled"

#### Scenario: No provider configured
- **WHEN** agent.multiAgent is true but agent.provider is empty
- **THEN** AgentRegistryCheck returns StatusFail indicating agent provider is not configured

### Requirement: Librarian health check
The doctor command SHALL include a LibrarianCheck that validates librarian configuration. The check SHALL skip when the librarian is disabled. When enabled, it SHALL verify that knowledge sources are configured.

#### Scenario: Librarian disabled
- **WHEN** doctor runs with librarian disabled
- **THEN** LibrarianCheck returns StatusSkip

#### Scenario: Librarian enabled with no knowledge sources
- **WHEN** librarian is enabled but no knowledge sources are configured
- **THEN** LibrarianCheck returns StatusWarn indicating no knowledge sources

### Requirement: Approval health check
The doctor command SHALL include an ApprovalCheck that validates approval system configuration. The check SHALL skip when the approval system is disabled. When enabled, it SHALL verify that the approval mode is valid and at least one approval channel is configured.

#### Scenario: Approval disabled
- **WHEN** doctor runs with approval system disabled
- **THEN** ApprovalCheck returns StatusSkip

#### Scenario: Approval enabled with valid configuration
- **WHEN** approval system is enabled with a valid mode and configured channel
- **THEN** ApprovalCheck returns StatusPass

### Requirement: New checks registered in AllChecks
The ToolHooksCheck, AgentRegistryCheck, LibrarianCheck, and ApprovalCheck SHALL be registered in the AllChecks() function under a "Tool Hooks / Agent Registry / Librarian / Approval" comment section.

#### Scenario: Doctor runs all new checks
- **WHEN** user runs `lango doctor`
- **THEN** the output includes results for "Tool Hooks", "Agent Registry", "Librarian", and "Approval" checks
