## Why

Lango's long-running tasks fail because there is no durable execution infrastructure. The agent "claims" completion but there is no system-level verification. This change introduces a Task OS that shifts from "agent claims completion" to "system proves completion."

Root cause: not memory, but lack of a durable execution engine with journal-based state, typed validators, and policy-driven failure recovery.

## What Changes

- New `internal/runledger/` package: append-only journal, materialized snapshots, PEV (Propose-Evidence-Verify) engine, 6 typed validators, policy supervisor protocol
- 8 new agent tools: `run_create`, `run_read`, `run_active`, `run_note`, `run_propose_step_result`, `run_apply_policy`, `run_approve_step`, `run_resume`
- Planner contract: strict JSON schema output with validation (ID uniqueness, DAG cycle detection, agent validation)
- Resume protocol: opt-in only, intent detection (Korean/English), staleness checking
- 3 new Ent schemas: `RunJournal`, `RunSnapshot`, `RunStep`
- Configuration: `runLedger.enabled`, shadow/writeThrough/authoritativeRead rollout stages
- Workspace isolation: coding-only, fail-closed git worktree management

## Capabilities

### New Capabilities
- `run-ledger`: Append-only journal-based durable execution engine
- `pev-engine`: Propose-Evidence-Verify validation with 6 typed validators
- `run-tools`: 8 agent tools for run lifecycle management
- `plan-parser`: Strict JSON schema parsing and validation for planner output
- `resume-protocol`: Opt-in resume with intent detection and staleness checking
- `workspace-isolation`: Git worktree-based coding isolation (fail-closed)
- `rollout-stages`: 4-stage progressive rollout (shadow → write-through → authoritative read → projection retired)

### Modified Capabilities
- `config-system`: Added RunLedgerConfig under root Config
- `appinit-modules`: Added ProvidesRunLedger key and runLedgerModule
- `app-types`: Added RunLedgerStore and RunLedgerPEV fields to App struct

## Impact

- **Code**: `internal/runledger/` (new, 12 files), `internal/ent/schema/` (3 new schemas), `internal/config/` (1 new file + 1 modified), `internal/app/` (1 new module + 2 modified), `internal/appinit/` (1 modified)
- **Dependencies**: No new external dependencies (uses existing `github.com/google/uuid`)
- **Config**: New `runLedger` section in lango.json with enabled, shadow, writeThrough, authoritativeRead, staleTtl, validatorTimeout, plannerMaxRetries
- **UX**: Agent can create durable runs, propose step results with PEV verification, and resume paused runs
