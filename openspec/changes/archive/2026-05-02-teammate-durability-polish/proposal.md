# Teammate Durability Polish

## Why

The first approval-blocked durability follow-up shipped the core mirror path, but the review left a few low-risk gaps open: the spec does not explicitly mention blocked-state replacement events, and the mirror decorator tests do not pin the negative no-op cases.

## What Changes

This change tightens the durability contract without changing user-facing behavior:

- document blocked-state replacement while a run remains approval-blocked
- clarify best-effort duplicate block events under concurrent projection writes
- add negative tests for no-op transition cases

## User Impact

No new feature is introduced. This is a correctness and maintenance polish pass on the existing approval-blocked durability mirror.
