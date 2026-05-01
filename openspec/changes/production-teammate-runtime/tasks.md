## 1. Rewrite Contracts

- [ ] 1.1 Rewrite `multi-agent-orchestration` contracts from static tool-less orchestration to dynamic teammate runtime semantics
- [ ] 1.2 Rewrite `agent-control-plane-tools` contracts for runtime-backed `agent_spawn`, `agent_wait`, and `agent_stop`
- [ ] 1.3 Rewrite `tool-execution-hooks`, `tool-capability-layer`, and `cli-agent-inspection` deltas for production teammate runtime behavior

## 2. Runtime Identity And Projection

- [ ] 2.1 Extend `AgentRun` projection with runtime identity chain, teammate type, projected condition, and visibility fields
- [ ] 2.2 Guarantee pre-registered `AgentRun.ID` reuse across background task execution and `RunLedger`
- [ ] 2.3 Surface projected non-terminal condition state from `agent_wait`

## 3. Spawn Contracts

- [ ] 3.1 Validate built-in teammate `allowed_tools` against role max scope during `agent_spawn`
- [ ] 3.2 Persist `spawn_reason` in the run projection and background submission payload
- [ ] 3.3 Submit spawned teammates through the existing in-process background manager path while preserving parent session context

## 4. Capability Policy

- [ ] 4.1 Preserve structured blocked-tool metadata through hook interception
- [ ] 4.2 Route blocked `DynamicAllowedTools` attempts into capability policy classification
- [ ] 4.3 Request approval only for in-scope tools and deny out-of-scope tools
- [ ] 4.4 Extend run `AllowedTools` exactly once after approval so the next run context carries the granted tool

## 5. Prompt And Inspection Surfaces

- [ ] 5.1 Update orchestration prompt guidance for dynamic teammate path and legacy fallback boundaries
- [ ] 5.2 Expose teammate runtime mode in `lango agent status` table and JSON output
- [ ] 5.3 Update CLI/docs surfaces affected by production teammate runtime behavior

## 6. Verification And Closeout

- [ ] 6.1 Run `openspec verify --change production-teammate-runtime`
- [ ] 6.2 Run `openspec status --change production-teammate-runtime` and `openspec instructions apply --change production-teammate-runtime`
- [ ] 6.3 Run `go build ./...`
- [ ] 6.4 Run `go test ./...`
- [ ] 6.5 Archive the change after implementation and downstream updates are complete
