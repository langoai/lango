## Overview

This change extends context-aware starter prompts with low-cost live Git signals.

## Design Decisions

### Git awareness remains opportunistic

The workbench only uses Git branch and dirty-state signals when:

- a repository root was already detected
- `git` is available in `PATH`
- the lightweight status queries succeed

If any of those checks fail, the prompt system falls back to the repository-aware behavior already in place.

### Only the change-review prompt changes

The repository summary and structure prompts remain stable. The third prompt becomes:

- branch-aware when the repo is clean
- branch-and-dirty aware when uncommitted changes are present

That keeps the quick-start set predictable while still making the most action-oriented prompt more useful.
