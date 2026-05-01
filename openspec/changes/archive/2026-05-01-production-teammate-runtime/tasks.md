## 1. Rewrite Contracts

- [x] 1.1 Rewrite `multi-agent-orchestration` contracts from static tool-less orchestration to dynamic teammate runtime semantics
- [x] 1.2 Rewrite `agent-control-plane-tools` contracts for runtime-backed `agent_spawn`, `agent_wait`, and `agent_stop`
- [x] 1.3 Rewrite `tool-execution-hooks`, `tool-capability-layer`, and `cli-agent-inspection` deltas for production teammate runtime behavior

## 2. Runtime Identity And Projection

- [x] 2.1 Extend `AgentRun` projection with runtime identity chain, teammate type, projected condition, and visibility fields
- [x] 2.2 Guarantee pre-registered `AgentRun.ID` reuse across background task execution and `RunLedger`
- [x] 2.3 Surface projected non-terminal condition state from `agent_wait`

## 3. Spawn Contracts

- [x] 3.1 Validate built-in teammate `allowed_tools` against role max scope during `agent_spawn`
- [x] 3.2 Persist `spawn_reason` in the run projection and background submission payload
- [x] 3.3 Submit spawned teammates through the existing in-process background manager path while preserving parent session context

## 4. Capability Policy

- [x] 4.1 Preserve structured blocked-tool metadata through hook interception
- [x] 4.2 Route blocked `DynamicAllowedTools` attempts into capability policy classification
- [x] 4.3 Request approval only for in-scope tools and deny out-of-scope tools
- [x] 4.4 Extend run `AllowedTools` exactly once after approval so the next run context carries the granted tool

## 5. Prompt And Inspection Surfaces

- [x] 5.1 Update orchestration prompt guidance for dynamic teammate path and legacy fallback boundaries
- [x] 5.2 Expose teammate runtime mode in `lango agent status` table and JSON output
- [x] 5.3 Update CLI/docs surfaces affected by production teammate runtime behavior

## 6. Verification And Closeout

- [x] 6.1 Run `openspec status --change production-teammate-runtime` and `openspec validate production-teammate-runtime --strict`
- [ ] 6.2 Archive the change through `openspec archive -y production-teammate-runtime`
- [x] 6.3 Run `go build ./...`
- [x] 6.4 Run `go test ./...`
- [ ] 6.5 Archive the change after implementation and downstream updates are complete
