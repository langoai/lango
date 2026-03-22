## ADDED Requirements

### Requirement: Shared automation helper package
The automation subsystem SHALL provide a shared `internal/automation/` helper package containing `AgentRunner`, `ChannelSender`, and `DetectChannelFromContext(ctx)` so cron, background, and workflow packages use one common contract.

#### Scenario: Automation packages share one runner interface
- **WHEN** `cron`, `background`, and `workflow` need to execute agent prompts
- **THEN** they SHALL depend on `automation.AgentRunner`
- **AND** they SHALL NOT define package-local duplicate `AgentRunner` interfaces

#### Scenario: Automation packages share one sender interface
- **WHEN** automation packages need channel delivery contracts
- **THEN** they SHALL depend on `automation.ChannelSender`
- **AND** they SHALL NOT define package-local duplicate `ChannelSender` interfaces

#### Scenario: Delivery target is derived from session context
- **WHEN** cron, background, or workflow tools omit an explicit delivery target
- **THEN** they SHALL use `automation.DetectChannelFromContext(ctx)` to derive `channel:target`
- **AND** they SHALL fall back to configured defaults only when no context-derived target exists
