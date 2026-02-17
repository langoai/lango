## Why

The Skill system was tightly coupled to SQLite/Ent ORM, creating unnecessary database overhead for what is essentially file-based configuration. Migrating to a `.lango/skills/` directory structure aligns with standard CLI tool patterns (like `.claude/skills/`) and removes the Ent Skill schema entirely, simplifying the persistence layer.

## What Changes

- **BREAKING**: Remove `SkillEntry` from `knowledge` package; moved to independent `skill` package
- **BREAKING**: Remove `knowledge.NewStore` `maxSkillsPerDay` parameter (4 args instead of 5)
- **BREAKING**: Remove `AutoApproveSkills` and `MaxSkillsPerDay` from `KnowledgeConfig`
- Add `SkillConfig` with `Enabled` and `SkillsDir` fields to root `Config`
- Implement `FileSkillStore` that reads/writes `<dir>/<name>/SKILL.md` files with YAML frontmatter
- Add SKILL.md parser (YAML frontmatter + markdown body)
- Embed 30 default CLI skills via `//go:embed` in `skills/` package
- Decouple `ContextRetriever` from skill storage via `SkillProvider` interface
- Delete `internal/ent/schema/skill.go` and regenerate Ent code
- Separate `initSkills()` from `initKnowledge()` in wiring layer

## Capabilities

### New Capabilities
- `file-based-skills`: File-based skill storage using `.lango/skills/<name>/SKILL.md` with YAML frontmatter parsing, embedded default skills, and `FileSkillStore` implementation

### Modified Capabilities

## Impact

- `internal/skill/` — new files: `types.go`, `store.go`, `parser.go`, `file_store.go`; modified: `registry.go`, `executor.go`, `builder.go`
- `internal/knowledge/` — `store.go` (removed skill methods), `types.go` (removed SkillEntry), `retriever.go` (added SkillProvider interface)
- `internal/app/` — `wiring.go` (separated initSkills), `app.go` (new init order), `tools.go` (updated meta-tools)
- `internal/config/` — `types.go` (added SkillConfig), `loader.go` (updated defaults)
- `internal/ent/schema/skill.go` — deleted, Ent regenerated
- `internal/learning/` — test files updated (NewStore signature change)
- `internal/cli/settings/` and `internal/cli/tuicore/` — removed skill config form fields
- `skills/` — new package with `embed.go` + 30 SKILL.md files
