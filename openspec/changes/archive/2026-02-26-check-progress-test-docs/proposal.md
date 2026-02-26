# Proposal: Test Coverage & Documentation Sync

## Problem

The dev branch contains ~31,400 lines of new P2P networking, security hardening, and sandbox code across 394 files. While the code architecture is solid, **critical packages lack test coverage** and **documentation has drifted from implementation**.

### Key Gaps
- **50+ source files** across 12 packages have zero test coverage
- `prompts/TOOL_USAGE.md` references a non-existent CLI command
- `docs/configuration.md` is missing ~10 configuration keys
- No integration tests for P2P discovery, identity, or CLI commands

## Proposed Solution

Add comprehensive unit tests for all untested packages, fix documentation inconsistencies, and ensure docs/config/code are fully aligned.

## Scope

- ~16 new test files covering P2P, CLI, workflow, cron, background, security, sandbox, librarian, payment
- 3 documentation files updated (TOOL_USAGE.md, configuration.md, p2p-network.md)
- OpenSpec change entry for tracking
