## ADDED Requirements

### Requirement: AgentDefinition type
The `agentregistry` package SHALL define an `AgentDefinition` struct with fields: Name, Description, Status, Capabilities, Prefixes, Keywords, AlwaysInclude, Instruction, Source, and metadata (Version, Author, Tags).

#### Scenario: AgentDefinition has all required fields
- **WHEN** an AgentDefinition is created
- **THEN** it SHALL have Name (string), Description (string), Status (string), Capabilities ([]string), Prefixes ([]string), Keywords ([]string), AlwaysInclude (bool), Instruction (string), and Source (AgentSource)

### Requirement: AgentSource enum
The package SHALL define an AgentSource enum with values: SourceBuiltin (0), SourceEmbedded (1), SourceUser (2), SourceRemote (3).

#### Scenario: AgentSource values
- **WHEN** AgentSource constants are referenced
- **THEN** SourceBuiltin SHALL be 0, SourceEmbedded SHALL be 1, SourceUser SHALL be 2, SourceRemote SHALL be 3

### Requirement: AGENT.md parser
The package SHALL provide a `ParseAgentMD` function that parses AGENT.md files with YAML frontmatter and markdown body. The YAML frontmatter SHALL contain structured metadata and the markdown body SHALL become the Instruction field.

#### Scenario: Parse valid AGENT.md
- **WHEN** a valid AGENT.md file with YAML frontmatter and markdown body is parsed
- **THEN** the YAML fields SHALL populate AgentDefinition metadata and the markdown body SHALL become the Instruction field

#### Scenario: Parse AGENT.md without frontmatter
- **WHEN** an AGENT.md file without YAML frontmatter is parsed
- **THEN** the parser SHALL return an error

#### Scenario: Roundtrip parsing
- **WHEN** an AgentDefinition is serialized to AGENT.md format and parsed back
- **THEN** all fields SHALL match the original definition

### Requirement: Registry with override semantics
The `Registry` SHALL support loading agents from multiple stores with override semantics: User overrides Embedded, Embedded overrides Builtin. An agent with the same name from a higher-priority source SHALL replace the lower-priority one.

#### Scenario: User overrides Embedded
- **WHEN** both embedded and user stores define an agent named "operator"
- **THEN** the Registry SHALL use the user-defined version

#### Scenario: Embedded overrides Builtin
- **WHEN** both builtin and embedded stores define an agent named "vault"
- **THEN** the Registry SHALL use the embedded version

### Requirement: Active agents filtering
The Registry SHALL provide an `Active()` method that returns only agents with status "active", sorted by name.

#### Scenario: Filter active agents
- **WHEN** the registry contains agents with status "active" and "disabled"
- **THEN** Active() SHALL return only "active" agents, sorted alphabetically by name

### Requirement: FileStore for user-defined agents
The `FileStore` SHALL load AGENT.md files from a directory structure: `<base>/<name>/AGENT.md`. Each subdirectory name SHALL become the agent name.

#### Scenario: Load from directory
- **WHEN** FileStore loads from a directory containing `operator/AGENT.md` and `custom/AGENT.md`
- **THEN** it SHALL return two AgentDefinitions with names "operator" and "custom" and Source set to SourceUser

### Requirement: EmbeddedStore for default agents
The `EmbeddedStore` SHALL load AGENT.md files from an `embed.FS` containing the 7 default agent definitions (operator, navigator, vault, librarian, automator, planner, chronicler).

#### Scenario: Load embedded defaults
- **WHEN** EmbeddedStore loads agents
- **THEN** it SHALL return 7 AgentDefinitions with Source set to SourceEmbedded

### Requirement: Store interface
The package SHALL define a `Store` interface with `Load() ([]AgentDefinition, error)` method. Both FileStore and EmbeddedStore SHALL implement this interface.

#### Scenario: Store implementations
- **WHEN** FileStore and EmbeddedStore are used
- **THEN** both SHALL implement the Store interface
