# Teammate Transient State Durability

## Why

The built-in multi-agent hard cut left one major durable visibility gap open: approval-blocked teammate state is truthful in the control-plane projection but not reconstructible from RunLedger after restart. The archived hard-cut audit marked `approval-blocked conditions` as `follow-up` for that reason in `openspec/changes/archive/2026-05-01-dynamic-multi-agent-hard-cut/design.md`.

## What Changes

This follow-up mirrors approval-blocked teammate transient state into RunLedger using journal events plus latest snapshot state. It covers `blocked_waiting_approval`, `blocked_reason`, and `grant_request_id` only. `recovery_state` remains follow-up because no production writer exists yet.

## User Impact

Operator-visible blocked teammate runs become durably inspectable later without changing the live read model. The runtime still treats RunLedger mirroring as best effort and does not fail-close if mirror writes fail.

Mirroring activates only when both `runLedger.enabled` and `runLedger.writeThrough` are set. Otherwise the runtime continues to rely on the in-memory projection only.
