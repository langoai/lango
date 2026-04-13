## Why

Pre-PR sync audit (dev → main) found 2 documentation/config inconsistencies. README.md hardcodes "80+ commands" while the actual count is 146, and mkdocs.yml nav is missing `cli/a2a.md` — the only cli/*.md file not included among 18 peers. These should be fixed before merge to maintain documentation quality.

## What Changes

- Replace hardcoded command count ("80+ commands") in README.md:119 with count-free phrasing to prevent future staleness
- Add `A2A Commands: cli/a2a.md` entry to mkdocs.yml CLI Reference nav section (all other 18 cli/*.md files are already listed)

## Capabilities

### New Capabilities

(none — documentation/config corrections only)

### Modified Capabilities

(none — no spec-level behavior changes)

## Impact

- `README.md`: User-facing landing page, command count accuracy improvement
- `mkdocs.yml`: Documentation site nav completeness, A2A CLI reference discoverability
- No code changes, no API changes, no dependency changes
