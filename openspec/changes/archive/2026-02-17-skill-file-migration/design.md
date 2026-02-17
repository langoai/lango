## Context

Lango's skill system was persisted via SQLite/Ent ORM alongside knowledge and learning entries. This created tight coupling between the skill lifecycle and the database layer. The migration replaces DB-backed skill storage with a file-based approach using `.lango/skills/<name>/SKILL.md` files with YAML frontmatter, matching standard CLI tool patterns.

## Goals / Non-Goals

**Goals:**
- Decouple skill storage from Ent ORM entirely
- Provide 30 embedded default CLI skills via `//go:embed`
- Maintain agent meta-tool compatibility (`create_skill`, `list_skills`)
- Simplify config by removing usage tracking and rate limiting for skills

**Non-Goals:**
- Migrating existing DB skill data (data is dropped)
- Changing the skill execution model (composite/script/template)
- Adding new skill types

## Decisions

- **File-based storage over DB**: Skills are configuration, not transactional data. File-based storage enables version control, manual editing, and aligns with `.claude/skills/` pattern. Alternative: keep DB but add file import — rejected due to unnecessary complexity.
- **YAML frontmatter + Markdown body**: Reuses a well-known format (Jekyll, Hugo). Alternatives: pure YAML, TOML — rejected because markdown body allows richer documentation.
- **`SkillProvider` interface for retriever decoupling**: Avoids import cycle between `knowledge` and `skill` packages. The adapter pattern (`skillProviderAdapter`) bridges `*skill.Registry` to `knowledge.SkillProvider`.
- **Embedded defaults via `//go:embed`**: Default skills ship with the binary. `FileSkillStore.EnsureDefaults()` copies them to user directory on first run without overwriting user modifications.
- **Drop usage tracking**: `UseCount`, `SuccessCount`, `LastUsedAt` removed. These were DB-only metrics with no consumer. Simplifies the `SkillEntry` type.

## Risks / Trade-offs

- **Data loss**: Existing DB-stored skills are not migrated. → Acceptable since the feature is early-stage with minimal user-created skills.
- **File I/O vs DB queries**: Listing skills requires directory scan. → Mitigated by small expected skill count (<100) and one-time load at startup.
- **No atomic writes**: File writes are not transactional. → Mitigated by write-then-rename pattern in `FileSkillStore.Save()`.
