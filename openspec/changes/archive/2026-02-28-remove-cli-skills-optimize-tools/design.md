## Context

The lango agent embeds 42 default skills in `skills/` — each a SKILL.md that wraps a `lango <command>` CLI call. At runtime these skills are deployed to the user's skills directory and registered as agent tools. However, every `lango` subprocess requires passphrase authentication during bootstrap, which always fails in non-interactive agent mode. The agent then wastes cycles attempting these broken skills before falling back to equivalent built-in tools.

## Goals / Non-Goals

**Goals:**
- Eliminate all 42 CLI wrapper SKILL.md files that always fail in agent mode
- Preserve the embed infrastructure so future (non-CLI) skills can still be embedded
- Add prompt-level guidance ensuring agents prefer built-in tools over skills
- Add runtime guidance in the knowledge retriever's skill section

**Non-Goals:**
- Changing the skill system architecture or SkillProvider interface
- Modifying how user-created or imported skills work
- Removing the `EnsureDefaults()` mechanism (it still works, just has no default skills to deploy)

## Decisions

### D1: Delete skill directories, keep embed.go with placeholder
**Decision:** Delete all 42 `skills/<name>/` directories. Preserve `skills/embed.go` with its original `//go:embed **/SKILL.md` directive. Add `skills/.placeholder/SKILL.md` to satisfy the glob pattern.
**Rationale:** The `go:embed` pattern errors at build time if zero files match. A placeholder file is simpler than rewriting embed.go or switching to a runtime filesystem approach. This keeps the door open for future embedded skills with zero code changes.
**Alternative considered:** Rewrite embed.go to return an empty `fs.FS` — rejected because it removes the embed pipeline entirely, requiring more work to restore later.

### D2: Multi-layer prompt guidance
**Decision:** Add tool priority notes at three levels: (1) `TOOL_USAGE.md` for detailed rules, (2) `AGENTS.md` for high-level directive, (3) `retriever.go` for runtime "Available Skills" section.
**Rationale:** LLM behavior is influenced by prompt reinforcement. A single mention may not reliably override the agent's tendency to try matching skills. Three layers (reference docs, identity prompt, runtime context) ensure consistent prioritization.

## Risks / Trade-offs

- **[Risk] Existing user scripts reference default skill names** → Mitigation: Skills were always wrappers around CLI commands that already fail. No working functionality is lost.
- **[Risk] Placeholder SKILL.md appears as a loadable skill** → Mitigation: The `.placeholder` directory name with leading dot is conventionally hidden and the SKILL.md lacks valid frontmatter, so `FileSkillStore` parsing will skip it.
- **[Trade-off]** Removing defaults reduces out-of-box discoverability of CLI commands via skill listing. Acceptable because the CLI `--help` system provides this discovery path.
