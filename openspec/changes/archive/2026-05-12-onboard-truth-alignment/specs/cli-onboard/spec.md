## ADDED Requirements

### Requirement: Onboard post-save output mentions all primary next steps
After saving, the onboard command SHALL display next-step guidance that mentions the default workbench entry point, the live runtime entry point, the verification command, and the full configuration editor.

#### Scenario: Post-save mentions primary entry points
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include `lango` as the default mission workbench entry point
- **AND** SHALL include `lango serve` as the live runtime entry point
- **AND** SHALL include `lango doctor` as the verification step
- **AND** SHALL include `lango settings` as the full-editor follow-up

### Requirement: Advanced feature guidance stays outside the five-step wizard
Product guidance for advanced systems SHALL describe them as settings-editor or config-import/export tasks instead of as dedicated onboarding submenu paths.

#### Scenario: Advanced features are routed to settings or config workflows
- **WHEN** a user needs to configure advanced systems such as embedding, graph, multi-agent, A2A, prompts, or OIDC providers
- **THEN** the product guidance SHALL direct them to `lango settings` or `lango config import/export`
- **AND** SHALL NOT describe nonexistent dedicated onboard submenu paths for those advanced systems
