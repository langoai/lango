## Context

The current runtime already documents the Chat key surface through `/help`, the status/help bar, and the approval-interrupt path. The missing artifact is the public cockpit feature reference, which should let operators understand the Chat page without reading the CLI overview or triggering `/help` first.

## Goals / Non-Goals

**Goals:**

- Document the Chat page as a first-class cockpit surface in the public feature reference.
- Keep the description concise but concrete about keys, slash commands, and approval controls.

**Non-Goals:**

- Change Chat runtime behavior.
- Duplicate every detail from the chat-specific CLI docs.

## Decisions

- Add one short `Chat Page` section with a compact keys table.
  - Rationale: the rest of the cockpit feature reference already uses dedicated page sections and key tables where they add operator value.

- Summarize slash commands at the level of discoverability rather than listing every command description in prose.
  - Rationale: the goal is navigation and expectation-setting, not replacing `/help`.

## Risks / Trade-offs

- [The feature doc may overlap slightly with CLI overview copy] → Keep the section focused on cockpit page behavior rather than generic chat-mode framing.
