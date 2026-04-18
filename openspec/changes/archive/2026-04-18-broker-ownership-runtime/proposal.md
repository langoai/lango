## Why

Bootstrap ownership has been partially brokered, but runtime domains still depend on direct stores and narrow raw-handle escape hatches. To reach true ownership, runtime readers and then mutators need to move onto broker-backed storage capabilities.

## What Changes

- Add broker-backed reader capabilities for learning history, librarian inquiries, workflow state, observability alerts, and reputation.
- Switch app and CLI runtime wiring to those broker-backed readers.
- Remove or retire remaining reader-side raw-handle escape hatches once the reader path is stable.

## Capabilities

### Modified Capabilities
- `brokered-storage`: runtime readers are served through broker-backed capabilities rather than in-process DB access.
- `cli-learning-inspection`, `cli-librarian-monitoring`, `cli-workflow-management`, `observability`, `p2p-reputation-cli`: move to broker-backed readers.

## Impact

- Affected code: `internal/storagebroker`, `internal/storage`, runtime app wiring, learning/librarian/workflow/observability/reputation CLI paths.
- This change is reader-first only. Payment/knowledge/agent-memory mutators remain for later phases in the same runtime ownership migration.
