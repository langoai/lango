## Context

Pre-PR sync audit (dev → main) found 2 documentation/config inconsistencies: README.md hardcodes a stale command count ("80+"), and mkdocs.yml CLI Reference nav omits `cli/a2a.md` while including all 18 other cli/*.md files.

## Goals / Non-Goals

**Goals:**
- Eliminate hardcoded command count that goes stale on every CLI addition
- Complete the mkdocs.yml CLI Reference nav so all cli/*.md files are listed

**Non-Goals:**
- Promoting other orphaned docs (cockpit guides, ADRs, research) — they are intentionally linked-only or internal
- Changing CI build tags — that is a separate policy decision

## Decisions

### Use count-free phrasing in README.md
Replace `"for all 80+ commands"` with `"for the complete command set"`. Updating to the correct number (146) would go stale again on the next CLI addition. Count-free phrasing eliminates recurring maintenance.

### Place A2A Commands after Agent & Memory in nav
`cli/a2a.md` is the only cli/*.md file missing from mkdocs.yml nav. A2A is agent-related functionality, so inserting after `Agent & Memory` follows the logical grouping.

## Risks / Trade-offs

No risks. Documentation/config-only changes, no code modifications.
