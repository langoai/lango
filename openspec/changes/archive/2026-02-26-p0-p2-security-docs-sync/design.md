## Context

All P0-P2 security hardening features are implemented in Go code but documentation (CLI docs, feature docs, README, agent prompts, security roadmap) has not been updated. Users cannot discover new features through `--help` cross-references, and the LLM agent lacks awareness of new P2P security capabilities in its system prompts.

## Goals / Non-Goals

**Goals:**
- Synchronize all documentation with the actual CLI implementation (exact flag names, output formats, JSON fields)
- Update agent prompts so the LLM knows about session management, sandbox, signed challenges, KMS, and credential revocation
- Mark all P0/P1 roadmap items as completed
- Ensure config key documentation matches `mapstructure` tags in `internal/config/types.go`

**Non-Goals:**
- No code changes — documentation-only
- No new features or behavioral changes
- No restructuring of existing documentation architecture

## Decisions

1. **Source-of-truth verification**: All CLI output examples are derived from actual `fmt.Printf`/`fmt.Println` calls in source files (status.go, keyring.go, kms.go, db_migrate.go, session.go, sandbox.go). JSON field names match `json` struct tags. Flag names match `cmd.Flags()` registrations.

2. **Documentation structure**: New CLI sections follow the existing pattern (Usage, Flags table, Example, JSON fields table). New feature sections follow existing heading hierarchy in each target file.

3. **Config table format**: New config rows in README follow the existing table format with key, type, default, and description columns. Deprecated fields are annotated inline.

## Risks / Trade-offs

- [Risk] Documentation may drift again as new features are added → Mitigation: OpenSpec workflow enforces documentation sync as part of change lifecycle
- [Risk] Large number of files modified increases merge conflict potential → Mitigation: All changes are additive (no deletions of existing content)
