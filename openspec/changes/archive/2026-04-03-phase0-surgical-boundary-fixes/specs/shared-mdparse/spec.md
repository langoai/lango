## ADDED Requirements

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
