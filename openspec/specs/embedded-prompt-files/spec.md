## Purpose

Production-quality default prompts stored as `.md` files and embedded into the binary via `go:embed`. Covers agent identity, safety rules, conversation behavior rules, and tool usage guidelines for Lango's 5 tool categories.

## Requirements

### Requirement: Prompt files embedded via go:embed
The system SHALL embed all default prompt `.md` files into the binary at build time using Go's `embed.FS`. The `prompts` package SHALL expose a public `FS` variable of type `embed.FS` containing all `.md` files in the package directory.

#### Scenario: All four prompt files are embedded
- **WHEN** the binary is built
- **THEN** `prompts.FS.ReadFile("AGENTS.md")` SHALL return non-empty content
- **AND** `prompts.FS.ReadFile("SAFETY.md")` SHALL return non-empty content
- **AND** `prompts.FS.ReadFile("CONVERSATION_RULES.md")` SHALL return non-empty content
- **AND** `prompts.FS.ReadFile("TOOL_USAGE.md")` SHALL return non-empty content

### Requirement: DefaultBuilder reads from embedded FS
`DefaultBuilder()` SHALL read prompt content from `prompts.FS` instead of hardcoded Go constants. The function signature and return type SHALL remain unchanged.

#### Scenario: Default builder uses embedded content
- **WHEN** `DefaultBuilder()` is called
- **THEN** the built prompt SHALL contain content from the embedded `.md` files
- **AND** the four sections (identity, safety, conversation_rules, tool_usage) SHALL be present

#### Scenario: Fallback on embed read failure
- **WHEN** `prompts.FS.ReadFile()` fails for any file
- **THEN** the system SHALL use a minimal fallback string for that section
- **AND** the system SHALL NOT panic or return an error

### Requirement: AGENTS.md covers agent identity
The `AGENTS.md` file SHALL define the agent's identity including name, role, eight tool categories (exec, filesystem, browser, crypto, secrets, cron, background, workflow), 6-layer knowledge system awareness, observational memory awareness, multi-channel awareness, and response principles.

#### Scenario: Identity prompt contains tool categories
- **WHEN** the identity section is rendered
- **THEN** it SHALL mention exec, filesystem, browser, crypto, secrets, cron, background, and workflow tools

#### Scenario: Identity prompt contains knowledge system
- **WHEN** the identity section is rendered
- **THEN** it SHALL reference the layered knowledge retrieval system

### Requirement: SAFETY.md covers security rules
The `SAFETY.md` file SHALL define security behavior rules including secret exposure prevention, destructive operation confirmation, PII protection, path traversal prevention, cryptographic material protection, environment variable filtering awareness, and ambiguous request verification.

#### Scenario: Safety prompt prevents secret exposure
- **WHEN** the safety section is rendered
- **THEN** it SHALL instruct never to expose secrets and reference the reference token system

#### Scenario: Safety prompt requires destructive operation confirmation
- **WHEN** the safety section is rendered
- **THEN** it SHALL require confirmation before destructive commands with specific examples

### Requirement: CONVERSATION_RULES.md prevents answer repetition
The `CONVERSATION_RULES.md` file SHALL define conversation behavior rules that prevent answer accumulation/repetition across turns, including current-question focus, self-contained responses, topic change handling, channel message limits, consistency maintenance, and channel-safe Markdown formatting conventions.

#### Scenario: Conversation rules prevent repetition
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL instruct to answer only the current question and not repeat previous answers

#### Scenario: Conversation rules respect channel limits
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL mention Telegram (4096) and Discord (2000) character limits

#### Scenario: Conversation rules include Markdown formatting conventions
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL instruct to use `-` for unordered lists instead of `*`
- **AND** it SHALL instruct to wrap code identifiers in backticks
- **AND** it SHALL instruct to ensure all formatting markers are paired
- **AND** it SHALL instruct to always close fenced code blocks

### Requirement: TOOL_USAGE.md provides per-tool guidelines
The `TOOL_USAGE.md` file SHALL provide specific usage guidelines for each of the eight tool categories with concrete patterns, commands, and best practices.

#### Scenario: Tool usage covers all eight tools
- **WHEN** the tool usage section is rendered
- **THEN** it SHALL contain subsections for Exec, Filesystem, Browser, Crypto, Secrets, Cron, Background, and Workflow tools

#### Scenario: Tool usage includes error handling guidance
- **WHEN** the tool usage section is rendered
- **THEN** it SHALL include guidance on handling tool errors and suggesting alternatives
