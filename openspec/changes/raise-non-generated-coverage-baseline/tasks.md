## 1. Planning And Baseline

- [x] 1.1 Measure the current non-Ent coverage baseline
- [x] 1.2 Identify top files by uncovered statement count
- [x] 1.3 Validate and commit this umbrella OpenSpec plan

## 2. Coverage Measurement Contract

- [x] 2.1 Add a repeatable coverage report command or script for non-generated Go code
- [x] 2.2 Exclude generated code by known generated paths and generated-file markers
- [x] 2.3 Add tests for the coverage filtering logic
- [x] 2.4 Report total coverage, covered statements, total statements, uncovered statements, and top uncovered files
- [x] 2.5 Add an executable threshold gate that fails when non-generated coverage is below 90%

## 3. Coverage Wave: Small Deterministic Packages

- [x] 3.1 Add focused tests for low-risk pure/helper packages with low coverage
- [x] 3.2 Add meaningful assertions for enum, parser, embed, and simple wrapper behavior
- [x] 3.3 Commit the wave separately

## 4. Coverage Wave: CLI/TUI State And Command Surfaces

- [x] 4.1 Add targeted tests for `internal/cli/tuicore/state_update.go`
- [x] 4.2 Add targeted tests for `internal/cli/settings/editor.go` and `internal/cli/settings/menu.go`
- [x] 4.3 Cover command construction, output seams, error paths, and state transitions
- [x] 4.4 Commit the wave separately

## 5. Coverage Wave: Storage And Knowledge

- [x] 5.1 Add focused tests for `internal/storage/facade.go`
- [x] 5.2 Add focused tests for `internal/storagebroker/server.go` and `internal/storagebroker/client.go`
- [x] 5.3 Add focused tests for `internal/knowledge/store.go`
- [x] 5.4 Cover CRUD, error handling, unavailable backends, and boundary conditions
- [x] 5.5 Commit the wave separately

## 6. Coverage Wave: Runtime And Workflow

- [x] 6.1 Add deterministic tests or seams for `internal/adk/agent.go`
- [x] 6.2 Add deterministic tests for `internal/workflow/engine.go`
- [x] 6.3 Add deterministic tests for `internal/turnrunner/runner.go`
- [x] 6.4 Add persistence boundary tests for `internal/turntrace/store.go`
- [x] 6.5 Cover workflow execution/status paths, turn-runner outcomes, and trace persistence boundaries
- [x] 6.6 Commit the wave separately

## 7. Coverage Wave: App Wiring And P2P

- [x] 7.1 Add focused tests for `internal/app/wiring_p2p.go`
- [x] 7.2 Add focused tests for `internal/app/wiring.go`
- [x] 7.3 Add focused tests for `internal/app/tools_meta.go`
- [x] 7.4 Add focused tests for `internal/p2p/handshake/handshake.go`
- [x] 7.5 Commit the wave separately

## 8. Coverage Wave: High-Impact Service Boundaries

- [x] 8.1 Add focused tests for remaining `internal/storagebroker/server.go` dispatch, payload protection, and error paths
- [x] 8.2 Add focused tests for remaining `internal/skill/importer.go` import validation, filesystem, and failure paths
- [x] 8.3 Add focused tests for remaining `internal/receipts/store.go` receipt lifecycle and query boundaries
- [x] 8.4 Commit the wave separately

## 9. Coverage Wave: App And CLI Execution Surfaces

- [x] 9.1 Add focused tests for remaining `internal/app/app.go`, `internal/app/modules.go`, and app wiring branches
- [x] 9.2 Add focused tests for `cmd/lango/main.go` command routing and error exits through seams
- [x] 9.3 Add focused tests for `internal/cli/chat/chat.go`, `internal/cli/settings/editor.go`, `internal/cli/sandbox/sandbox.go`, and `internal/cli/onboard/wizard.go`
- [x] 9.4 Commit the wave separately

## 10. Coverage Wave: Remaining P2P And Payment Surfaces

- [x] 10.1 Add focused tests for remaining `internal/p2p/handshake/handshake.go` stream and approval branches
- [x] 10.2 Add focused tests for high-uncovered P2P protocol, settlement, discovery, and payment paths that do not require external services
- [x] 10.3 Add focused tests for remaining `internal/app/tools_escrow.go` and payment tool wrappers
- [x] 10.4 Commit the wave separately

## 11. Coverage Wave: Cron, Broker, And App Helpers

- [x] 11.1 Add focused tests for `internal/cron/store.go` Ent-backed store and conversion helpers
- [x] 11.2 Add focused tests for remaining `internal/storagebroker/server.go` and `internal/storagebroker/client.go` RPC wrapper branches
- [x] 11.3 Add focused tests for app P2P helper surfaces that do not require external services
- [x] 11.4 Commit the wave separately

## 12. Coverage Wave: Smart Account, Wallet, And Git Bundle

- [x] 12.1 Add focused tests for `internal/wallet/spending.go` limiter behavior and store error paths
- [x] 12.2 Add focused tests for `internal/smartaccount/manager.go` manager helpers and encoding branches
- [x] 12.3 Add focused tests for `internal/p2p/gitbundle/protocol.go` protocol handler dispatch and error branches
- [x] 12.4 Add focused tests for app Smart Account tools and wiring helper branches
- [x] 12.5 Commit the wave separately

## 13. Coverage Wave: Filesystem, Test Utilities, And Knowledge Branches

- [x] 13.1 Add focused tests for `internal/tools/filesystem/filesystem.go` delete, exists, mkdir, copy, and path access branches
- [x] 13.2 Add focused tests for `internal/testutil/mock_session_store.go` CRUD, salt, end, counters, copy semantics, and error injection
- [x] 13.3 Add focused tests for `internal/knowledge/store.go` FTS5 result resolution, learning lookup branches, and low-risk search paths
- [x] 13.4 Commit the wave separately

## 14. Coverage Wave: App P2P, Meta Tools, And Knowledge Store

- [x] 14.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` helper and wiring branches
- [x] 14.2 Add focused tests for deterministic `internal/app/tools_meta.go` tool handlers and helper branches
- [x] 14.3 Add focused tests for remaining low-risk `internal/knowledge/store.go` scoring, stats, external ref, and deletion branches
- [x] 14.4 Run subagent-driven review checkpoints for the wave
- [x] 14.5 Commit the wave separately

## 15. Coverage Wave: CLI Entrypoints, Chat State, Settings Navigation, And Config Redaction

- [x] 15.1 Add focused tests for deterministic `cmd/lango/main.go` helper and adapter branches
- [x] 15.2 Add focused tests for deterministic `internal/cli/chat/chat.go` state, accessor, session-mode, and key-helper branches
- [x] 15.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` redaction, scalar, reflection, formatting, and set-field branches
- [x] 15.4 Add focused tests for deterministic `internal/cli/settings/editor.go` dependency navigation and setup-flow branches
- [x] 15.5 Run subagent-driven review checkpoints for the wave
- [x] 15.6 Commit the wave separately

## 16. Coverage Wave: Gateway, Telegram, And App Adapters

- [x] 16.1 Add focused tests for deterministic `internal/gateway/server.go` setter, status, RPC, and companion branches
- [x] 16.2 Add focused tests for deterministic `internal/channels/telegram/telegram.go` channel identity, send, split, error formatting, and allow-list branches
- [x] 16.3 Add focused tests for deterministic `internal/app/app.go` publish, post-agent wiring, and lifecycle registration branches
- [x] 16.4 Run subagent-driven review checkpoints for the wave
- [x] 16.5 Commit the wave separately

## 17. Coverage Wave: App Wiring, Knowledge Store, And CLI/ADK Surfaces

- [x] 17.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/tools_meta.go`, `internal/app/modules.go`, and related app wiring branches
- [x] 17.2 Add focused tests for deterministic `internal/knowledge/store.go` search, FTS5, payload-protection, deletion, and learning branches
- [x] 17.3 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/chat/chat.go`, and `internal/adk` option/helper branches
- [x] 17.4 Run subagent-driven review checkpoints for the wave
- [x] 17.5 Commit the wave separately

## 18. Coverage Wave: Integrations, Ledgers, And Receipts

- [x] 18.1 Add focused tests for deterministic `internal/channels/discord/discord.go` channel identity, send, split, guild allow-list, and error formatting branches
- [x] 18.2 Add focused tests for deterministic `internal/p2p/gitbundle/bundle.go` diff, apply, hook, and safe-apply branches
- [x] 18.3 Add focused tests for deterministic `internal/runledger/writethrough.go`, `internal/mcp/connection.go`, and `internal/receipts/store.go` branch-heavy helpers
- [x] 18.4 Run subagent-driven review checkpoints for the wave
- [x] 18.5 Commit the wave separately

## 19. Coverage Wave: App, Knowledge, And User-Facing Entrypoints

- [x] 19.1 Add focused tests for deterministic `internal/app` metadata, module, wiring, and app helper branches that do not require live network services
- [x] 19.2 Add focused tests for deterministic `internal/knowledge/store.go` persistence, validation, and branch-heavy query helpers
- [x] 19.3 Add focused tests for deterministic `cmd/lango/main.go` and high-uncovered user-facing CLI helper branches without launching interactive UI loops
- [x] 19.4 Run subagent-driven review checkpoints for the wave
- [x] 19.5 Commit the wave separately

## 20. Coverage Wave: CLI TUI Pages, Settings, And ADK Branches

- [x] 20.1 Add focused tests for deterministic `internal/cli/cockpit/pages/missioncontrol.go` rendering, navigation, filter, and command-creation branches
- [x] 20.2 Add focused tests for deterministic `internal/cli/settings/editor.go` editor state, section navigation, validation, and save/cancel helper branches
- [x] 20.3 Add focused tests for deterministic `internal/adk/agent.go` run helper, collection, diagnostics, and error branch helpers without live provider calls
- [x] 20.4 Run subagent-driven review checkpoints for the wave
- [x] 20.5 Commit the wave separately

## 21. Coverage Wave: Knowledge Atomic Branches And App Wiring Hotspots

- [x] 21.1 Add focused tests for deterministic `internal/knowledge/store.go` atomic payload, FTS5 sync, and error helper branches without relying on generated code coverage
- [x] 21.2 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool handler, dispatcher, and receipt/error branches without live network services
- [x] 21.3 Add focused tests for deterministic `internal/app/modules.go`, `internal/app/app.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` module/wiring helper branches
- [x] 21.4 Run subagent-driven review checkpoints for the wave
- [x] 21.5 Commit the wave separately

## 22. Coverage Wave: CLI Entrypoint, Sandbox Probe, And Skill Importer Branches

- [x] 22.1 Add focused tests for deterministic `cmd/lango/main.go` chat/cockpit/workbench/config command branch helpers without launching interactive UI loops
- [x] 22.2 Add focused tests for deterministic `internal/cli/sandbox/sandbox.go` status collection, test/probe command helpers, and sandbox capability branches without requiring privileged sandbox execution
- [x] 22.3 Add focused tests for deterministic `internal/skill/importer.go` git/http import planning, URL/resource handling, and failure branches without live network access
- [x] 22.4 Run subagent-driven review checkpoints for the wave
- [x] 22.5 Commit the wave separately

## 23. Coverage Wave: Escrow Tools, Economy Wiring, And Chat Approval Branches

- [x] 23.1 Add focused tests for deterministic `internal/app/tools_escrow.go` escrow tool handler success, validation, and error branches without live chain calls
- [x] 23.2 Add focused tests for deterministic `internal/app/wiring_economy.go`, `internal/app/wiring.go`, and related app wiring helper branches without live providers or network services
- [x] 23.3 Add focused tests for deterministic `internal/cli/chat/chat.go` approval/key handling, pending state, and submission guard branches without launching the TUI program
- [x] 23.4 Run subagent-driven review checkpoints for the wave
- [x] 23.5 Commit the wave separately

## 24. Coverage Wave: App Meta, Module Initialization, P2P Wiring, And CLI Entrypoints

- [x] 24.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool composition, dependency-disabled, and handler edge branches without live network services
- [x] 24.2 Add focused tests for deterministic `internal/app/modules.go` module initialization, catalog, status, and disabled-dependency branches without live providers
- [x] 24.3 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, and `cmd/lango/main.go` startup and wiring helper branches without binding external services
- [x] 24.4 Run subagent-driven review checkpoints for the wave
- [x] 24.5 Commit the wave separately

## 25. Coverage Wave: Contract Caller, A2A Card, Cover CLI, And ADK Context Helpers

- [x] 25.1 Add focused tests for deterministic `cmd/lango-cover/main.go` argument, module discovery, threshold, and error branches without external commands
- [x] 25.2 Add focused tests for deterministic `internal/a2a/server.go` agent-card construction, P2P/pricing mutation, route output, and skill extraction branches without network listeners
- [x] 25.3 Add focused tests for deterministic `internal/contract/caller.go` ABI, revert-reason, receipt timeout/context, and local error branches without live RPC services
- [x] 25.4 Add focused tests for deterministic `internal/adk/context_model.go` context item, token estimation, and retrieval merge helper branches without live providers
- [x] 25.5 Run subagent-driven review checkpoints for the wave
- [x] 25.6 Commit the wave separately

## 26. Coverage Wave: Mission Projector, Ontology Service, And App/Main Helper Branches

- [x] 26.1 Add focused tests for deterministic `internal/cli/cockpit/missioncontrol_projector.go` projection helper, status mapping, summary, and enrichment branches without launching the TUI
- [x] 26.2 Add focused tests for deterministic `internal/ontology/service.go` action executor, schema health, type usage, permission, and delegation error branches without external services
- [x] 26.3 Add focused tests for deterministic `internal/app` and `cmd/lango/main.go` helper branches that remain high-uncovered without binding listeners or launching interactive loops
- [x] 26.4 Run subagent-driven review checkpoints for the wave
- [x] 26.5 Commit the wave separately

## 27. Coverage Wave: Meta Tools, Receipts Store, Ontology Tools, And App Wiring Hotspots

- [x] 27.1 Add focused tests for deterministic `internal/app/tools_meta.go` and `internal/app/tools_escrow.go` validation, dispatcher, receipt, and helper branches without live network or chain calls
- [x] 27.2 Add focused tests for deterministic `internal/receipts/store.go` lifecycle, query, filtering, and error branches using in-memory or temp-file state only
- [x] 27.3 Add focused tests for deterministic `internal/ontology/tools.go` tool builder, parameter validation, action/query, and error branches without external services
- [x] 27.4 Add focused tests for deterministic `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` helper branches without binding listeners or starting P2P services
- [x] 27.5 Run subagent-driven review checkpoints for the wave
- [x] 27.6 Commit the wave separately

## 28. Coverage Wave: Gateway, P2P Protocol, Gemini Provider, And Remaining App Entrypoints

- [x] 28.1 Add focused tests for deterministic `internal/gateway/server.go` response handlers, callbacks, router setup, and connection lifecycle branches without binding real listeners
- [x] 28.2 Add focused tests for deterministic `internal/p2p/protocol/remote_agent.go` request/response, timeout, encoding, and error branches with fake streams/hosts only
- [x] 28.3 Add focused tests for deterministic `internal/provider/gemini/gemini.go` provider identity, schema conversion, request assembly, model listing, and error branches without live Gemini calls
- [x] 28.4 Add focused tests for deterministic `internal/app/tools_p2p.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `cmd/lango/main.go` helper branches without starting P2P services or interactive loops
- [x] 28.5 Run subagent-driven review checkpoints for the wave
- [x] 28.6 Commit the wave separately

## 29. Coverage Wave: App Metadata, Entrypoints, Test Utilities, And Economy Hotspots

- [x] 29.1 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` branches that remain high-uncovered without live providers or listeners
- [x] 29.2 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/settings/forms_economy.go`, and `internal/testutil/mock_cron.go` helper branches without launching interactive UI loops or relying on wall-clock scheduling
- [x] 29.3 Add focused tests for deterministic `internal/adk/agent.go`, `internal/bootstrap/phases.go`, and `internal/economy/tools.go` branch-heavy helpers without live model providers, chain RPC, or external services
- [x] 29.4 Run subagent-driven review checkpoints for the wave
- [x] 29.5 Commit the wave separately

## 30. Coverage Wave: Remaining App Hotspots, TUI Form, Smart Account, And Workflow CLI

- [x] 30.1 Add focused tests for remaining deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` branches that still dominate uncovered statement count without live providers, listeners, or chain RPC
- [x] 30.2 Add focused tests for deterministic `internal/cli/tuicore/form.go`, `internal/cli/smartaccount/session.go`, and `internal/cli/workflow/workflow.go` helper branches without launching interactive UI loops
- [x] 30.3 Add focused tests for remaining deterministic `cmd/lango/main.go` and `internal/adk/agent.go` helper branches that are safe to exercise with fake stores, fake agents, and no external services
- [x] 30.4 Run subagent-driven review checkpoints for the wave
- [x] 30.5 Commit the wave separately

## 31. Coverage Wave: Status, Settings Knowledge, P2P Node, And ADK Support Helpers

- [x] 31.1 Add focused tests for deterministic `internal/cli/settings/forms_knowledge.go` form builders and field wiring without launching interactive UI loops
- [x] 31.2 Add focused tests for deterministic `internal/cli/status/status.go` dead-letter bridge, retry/detail, sanitization, and status helper branches using fake stores and command writers
- [x] 31.3 Add focused tests for deterministic `internal/p2p/node.go` constructor/accessor, key path, handler, and peer-found helper branches without external peers
- [x] 31.4 Add focused tests for deterministic `internal/adk` support helpers such as child session service, context helpers, PII adapter, plugin constructors, and tool adapters without live providers
- [x] 31.5 Run subagent-driven review checkpoints for the wave
- [x] 31.6 Commit the wave separately

## 32. Coverage Wave: Automation Settings Forms And Smart Account CLI Branches

- [x] 32.1 Add focused tests for deterministic `internal/cli/settings/forms_automation.go` form builders, field wiring, defaults, and validation branches without launching interactive UI loops
- [x] 32.2 Add focused tests for deterministic `internal/cli/smartaccount/policy.go` and remaining `internal/cli/smartaccount/session.go` command output, cleanup, validation, and error branches without live chain RPC
- [x] 32.3 Run subagent-driven review checkpoints for the wave
- [x] 32.4 Commit the wave separately

## 33. Coverage Wave: Session Store, Storage Broker, App Runtime Helpers

- [x] 33.1 Add focused tests for deterministic `internal/session/ent_store.go` constructor, client/DB accessor, list, timeout annotation, checksum, and migration branches using temp SQLite state only
- [x] 33.2 Add focused tests for deterministic `internal/storagebroker/server.go` dispatch, payload encode/decode, crypto wrapper, and unavailable dependency branches without binding real listeners
- [x] 33.3 Add focused tests for remaining deterministic `internal/app` and `cmd/lango/main.go` runtime helper branches without live providers, external listeners, P2P services, or interactive UI loops
- [x] 33.4 Run subagent-driven review checkpoints for the wave
- [x] 33.5 Commit the wave separately

## 34. Coverage Wave: App Meta, Network, And Wiring Branches

- [x] 34.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool construction and exportability/payment helper branches without live payments, chain RPC, or background workers
- [x] 34.2 Add focused tests for deterministic `internal/app/modules.go` network/intelligence module disabled and lightweight catalog branches without starting P2P, payment, or external services
- [x] 34.3 Add focused tests for deterministic `internal/app/wiring.go` and `internal/app/wiring_p2p.go` helper branches such as security/provenance/agent options and invalid or disabled P2P configuration paths
- [x] 34.4 Run subagent-driven review checkpoints for the wave
- [ ] 34.5 Commit the wave separately

## 35. Verification And Enforcement

- [ ] 35.1 Run `go build ./...`
- [ ] 35.2 Run `go test ./...`
- [ ] 35.3 Run the non-generated coverage report and confirm coverage is at least 90%
- [ ] 35.4 Run the executable 90% coverage gate and confirm it passes
- [ ] 35.5 Run `git diff --check`
- [ ] 35.6 Run subagent-driven spec and code-quality review checkpoints
- [ ] 35.7 Run `openspec validate --all --strict`
- [ ] 35.8 Archive the completed OpenSpec change
