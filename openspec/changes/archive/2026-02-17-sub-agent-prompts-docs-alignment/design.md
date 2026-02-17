## Context

Lango's multi-agent system has 7 sub-agents defined in `internal/orchestration/tools.go` via `agentSpecs`. The per-agent prompt customization system (`internal/prompt/builder.go`) loads `agents/<name>/IDENTITY.md` files from a prompts directory, but no default IDENTITY.md files existed in `prompts/agents/`. Additionally, the automator sub-agent (added with automation tools) was missing from global prompts, README documentation, and the docker-compose.yml lacked a runtime prompt customization option.

## Goals / Non-Goals

**Goals:**
- Provide default IDENTITY.md files for all 7 sub-agents so users have a reference when customizing prompts
- Document automation tools (Cron, Background, Workflow) in global prompts alongside existing tool categories
- Ensure README accurately reflects all 7 sub-agents in every reference location
- Enable runtime prompt customization in Docker via an optional volume mount

**Non-Goals:**
- Changing Go code or agent behavior (prompt content matches existing `agentSpecs[].Instruction` exactly)
- Modifying the prompt loading mechanism in `internal/prompt/builder.go`
- Adding new sub-agents or tools

## Decisions

1. **IDENTITY.md content mirrors agentSpecs[].Instruction verbatim**: Keeps a single source of truth. The Go code remains authoritative; the files serve as user-facing defaults and documentation.

2. **Flat file structure `prompts/agents/<name>/IDENTITY.md`**: Matches the existing convention expected by `PromptBuilder.loadAgentPrompts()`. No new directory patterns needed.

3. **Commented volume mount in docker-compose.yml**: Non-breaking change. Users opt-in by uncommenting. The existing `COPY prompts/` in Dockerfile already includes the new files automatically.

4. **Automation tool docs follow existing section patterns**: Cron/Background/Workflow sections in TOOL_USAGE.md follow the same format as Exec/Filesystem/Browser/Crypto/Secrets.

## Risks / Trade-offs

- [Content drift] The IDENTITY.md files may diverge from `agentSpecs[].Instruction` over time if one is updated without the other → Mitigation: Document that Go code is authoritative; files are defaults for user reference.
- [No automated sync] There is no mechanism to auto-generate IDENTITY.md from Go code → Mitigation: Acceptable for 7 files; could add a codegen step later if needed.
