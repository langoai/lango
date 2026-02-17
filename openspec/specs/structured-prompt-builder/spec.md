## Purpose

Section-based system prompt construction with priority ordering, default sections, and file-based override/customization. Resolves LLM answer repetition by including conversation behavior rules by default.

## Requirements

### Requirement: PromptSection interface
The system SHALL define a `PromptSection` interface with `ID() SectionID`, `Priority() int`, and `Render() string` methods. Sections with empty `Render()` output SHALL be omitted from the final prompt.

#### Scenario: Section renders with title
- **WHEN** a StaticSection has a non-empty title and content
- **THEN** Render() SHALL return "## {title}\n{content}"

#### Scenario: Section renders without title
- **WHEN** a StaticSection has an empty title and non-empty content
- **THEN** Render() SHALL return the content only

#### Scenario: Empty section is omitted
- **WHEN** a StaticSection has empty or whitespace-only content
- **THEN** Render() SHALL return an empty string

### Requirement: Builder assembles sections by priority
The system SHALL provide a Builder that collects PromptSection instances, sorts them by priority (lower first), and joins rendered output with double newlines.

#### Scenario: Priority ordering
- **WHEN** sections with priorities 300, 100, 200 are added
- **THEN** Build() SHALL output them in order 100, 200, 300

#### Scenario: Same-ID replacement
- **WHEN** two sections with the same ID are added
- **THEN** the last-added section SHALL replace the earlier one

#### Scenario: Section removal
- **WHEN** a section is removed by ID
- **THEN** Build() SHALL not include that section in the output

### Requirement: Default sections included
The system SHALL provide a `DefaultBuilder()` that returns a Builder with four built-in sections: Identity (priority 100), Safety (priority 200), Conversation Rules (priority 300), and Tool Usage (priority 400). Section content SHALL be read from embedded `.md` files in the `prompts` package via `prompts.FS.ReadFile()`. If an embedded file read fails, the system SHALL use a minimal fallback string for that section.

#### Scenario: Default builder includes conversation rules
- **WHEN** DefaultBuilder().Build() is called
- **THEN** the output SHALL contain conversation rules instructing the LLM to focus on the current question and not repeat previous answers

#### Scenario: Default builder section order
- **WHEN** DefaultBuilder().Build() is called
- **THEN** Identity SHALL appear before Safety, Safety before Conversation Rules, Conversation Rules before Tool Usage

#### Scenario: Default builder uses embedded content
- **WHEN** DefaultBuilder().Build() is called with a correctly built binary
- **THEN** the Identity section SHALL contain content from AGENTS.md describing Lango's five tool categories
- **AND** the Safety section SHALL contain content from SAFETY.md with security rules
- **AND** the Conversation Rules section SHALL contain content from CONVERSATION_RULES.md
- **AND** the Tool Usage section SHALL contain content from TOOL_USAGE.md with per-tool guidelines

### Requirement: Directory-based prompt loading
The system SHALL provide a `LoadFromDir` function that reads `.md` files from a directory and overrides matching default sections. Known filenames (AGENTS.md, SAFETY.md, CONVERSATION_RULES.md, TOOL_USAGE.md) SHALL map to their respective section IDs. Unknown `.md` files SHALL be added as custom sections with priority 900+.

#### Scenario: Override known section via file
- **WHEN** an AGENTS.md file exists in the prompts directory with content "Custom identity"
- **THEN** the Identity section SHALL contain "Custom identity" instead of the default

#### Scenario: Add custom section from unknown file
- **WHEN** a MY_RULES.md file exists in the prompts directory
- **THEN** it SHALL be added as a custom section with priority >= 900

#### Scenario: Empty file does not override
- **WHEN** a known filename exists but contains only whitespace
- **THEN** the default section SHALL remain unchanged

#### Scenario: Non-existent directory falls back to defaults
- **WHEN** LoadFromDir is called with a non-existent directory path
- **THEN** the builder SHALL return with all default sections intact

### Requirement: Builder Clone method
The prompt Builder SHALL provide a `Clone()` method that returns a deep copy of the builder, allowing independent modification without affecting the original.

#### Scenario: Clone produces independent copy
- **WHEN** a builder is cloned and the clone is modified (Add/Remove)
- **THEN** the original builder's sections SHALL remain unchanged

#### Scenario: Clone of empty builder
- **WHEN** an empty builder is cloned
- **THEN** the clone SHALL be a valid empty builder that can accept new sections

### Requirement: SectionAgentIdentity constant
The prompt package SHALL define a `SectionAgentIdentity` section ID with the value `"agent_identity"`, used for sub-agent role descriptions at priority 150.

#### Scenario: SectionAgentIdentity is distinct from SectionIdentity
- **WHEN** both SectionIdentity and SectionAgentIdentity are added to a builder
- **THEN** they SHALL coexist as separate sections (not replace each other)

### Requirement: LoadAgentFromDir function
The prompt package SHALL provide a `LoadAgentFromDir(base *Builder, dir string, logger)` function that overlays per-agent prompt overrides from a directory onto a cloned base builder.

#### Scenario: Known file mapping
- **WHEN** `IDENTITY.md` exists in the agent directory
- **THEN** it SHALL be loaded as SectionAgentIdentity with priority 150

#### Scenario: Empty or missing files ignored
- **WHEN** a `.md` file in the agent directory is empty or whitespace-only
- **THEN** it SHALL be ignored and the base section SHALL remain

#### Scenario: Non-existent directory returns base
- **WHEN** the agent directory does not exist
- **THEN** the original base builder SHALL be returned unmodified
