## MODIFIED Requirements

### Requirement: Tool category documentation in AGENTS.md
The AGENTS.md prompt file SHALL document all available tool categories. In addition to the existing 5 categories (Exec, Filesystem, Browser, Crypto, Secrets), it SHALL include 3 automation categories: Cron, Background, and Workflow.

#### Scenario: Automation categories present in AGENTS.md
- **WHEN** AGENTS.md is loaded as the agent identity prompt
- **THEN** it SHALL contain descriptions for Cron, Background, and Workflow tool categories alongside the existing 5 categories

### Requirement: Tool usage guidelines in TOOL_USAGE.md
The TOOL_USAGE.md prompt file SHALL provide usage guidelines for all tool types. In addition to existing sections (Exec, Filesystem, Browser, Crypto, Secrets, Tool Approval, Error Handling), it SHALL include sections for Cron Tool, Background Tool, and Workflow Tool.

#### Scenario: Cron Tool section in TOOL_USAGE.md
- **WHEN** TOOL_USAGE.md is loaded
- **THEN** it SHALL contain a Cron Tool section describing cron_add, cron_list, cron_pause, cron_resume, cron_remove, and cron_history usage

#### Scenario: Background Tool section in TOOL_USAGE.md
- **WHEN** TOOL_USAGE.md is loaded
- **THEN** it SHALL contain a Background Tool section describing bg_submit, bg_status, bg_list, and bg_result usage

#### Scenario: Workflow Tool section in TOOL_USAGE.md
- **WHEN** TOOL_USAGE.md is loaded
- **THEN** it SHALL contain a Workflow Tool section describing workflow_run, workflow_save, workflow_status, and workflow_cancel usage
