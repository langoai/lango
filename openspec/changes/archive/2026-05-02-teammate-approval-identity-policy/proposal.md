# Teammate Approval Identity Policy

## Why

Built-in teammate approval blocking now has durable visibility, but the approval identity itself is still underspecified. `GrantRequestID` is currently deterministic (`grant-{runID}-{toolName}`), which keeps the runtime simple but leaves two operator-facing questions unanswered:

- what identity should a second approval request for the same run and tool use?
- how should rejection, retry, and deduplication behave across UI, runtime, and durable state?

Without a clear contract, the runtime, control-plane tools, and any approval surface can drift in incompatible ways.

## What Changes

This change defines the identity and lifecycle policy for teammate approval requests:

- `GrantRequestID` remains a stable logical request identity per `(run, tool)`
- how repeated requests for the same `(run, tool)` are represented
- how denial and re-request semantics work
- what live and durable surfaces must expose

## User Impact

Approval-blocked teammate runs become easier to reason about operationally. A user or operator can tell whether a later request is the same approval identity or a new attempt, and tooling can deduplicate or display the request consistently.
