### Requirement: Shared frontmatter parser
The `mdparse` package SHALL provide a `SplitFrontmatter(content []byte) ([]byte, string, error)` function that extracts YAML frontmatter and body from markdown content with `---` delimiters.

#### Scenario: Valid frontmatter extraction
- **WHEN** content starts with `---`, followed by YAML, then a closing `---`, then body text
- **THEN** `SplitFrontmatter` SHALL return the YAML bytes, trimmed body string, and nil error

#### Scenario: Missing opening delimiter
- **WHEN** content does not start with `---`
- **THEN** `SplitFrontmatter` SHALL return an error containing "missing frontmatter delimiter"

#### Scenario: Missing closing delimiter
- **WHEN** content starts with `---` but has no closing `---`
- **THEN** `SplitFrontmatter` SHALL return an error containing "missing closing frontmatter delimiter"

### Requirement: Skill parser delegates to mdparse
The `skill` package's `splitFrontmatter` SHALL delegate to `mdparse.SplitFrontmatter` instead of implementing its own copy.

#### Scenario: Skill parser uses shared implementation
- **WHEN** `ParseSkillMD` is called with valid SKILL.md content
- **THEN** the frontmatter extraction SHALL be performed by `mdparse.SplitFrontmatter`

### Requirement: Agent registry parser delegates to mdparse
The `agentregistry` package's `splitFrontmatter` SHALL delegate to `mdparse.SplitFrontmatter` instead of implementing its own copy.

#### Scenario: Agent registry parser uses shared implementation
- **WHEN** `ParseAgentMD` is called with valid AGENT.md content
- **THEN** the frontmatter extraction SHALL be performed by `mdparse.SplitFrontmatter`

### Requirement: mdparse provides RenderFrontmatter
The `internal/mdparse` package SHALL export `RenderFrontmatter(meta interface{}, body string) ([]byte, error)` that produces the standard frontmatter format: `---\n(YAML)\n---\n\n(body)`.

#### Scenario: Struct metadata is rendered
- **WHEN** `RenderFrontmatter(struct{Name string}{"test"}, "body text")` is called
- **THEN** the output is `---\nname: test\n---\n\nbody text`

#### Scenario: Roundtrip with SplitFrontmatter
- **WHEN** `RenderFrontmatter(meta, body)` output is passed to `SplitFrontmatter`
- **THEN** the parsed frontmatter and body match the original inputs

### Requirement: agentregistry uses mdparse.RenderFrontmatter
The `internal/agentregistry/parser.go` SHALL delegate frontmatter rendering to `mdparse.RenderFrontmatter` instead of inline `---\n` + yaml.Marshal + `---\n\n` construction.

#### Scenario: RenderAgentMD output is unchanged
- **WHEN** `RenderAgentMD` is called with the same inputs as before
- **THEN** the output is byte-identical to the previous inline implementation

### Requirement: skill uses mdparse.RenderFrontmatter
The `internal/skill/parser.go` SHALL delegate frontmatter rendering to `mdparse.RenderFrontmatter` instead of inline construction.

#### Scenario: RenderSkillMD output is unchanged
- **WHEN** `RenderSkillMD` is called with the same inputs as before
- **THEN** the output is byte-identical to the previous inline implementation
