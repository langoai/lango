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

## 3. Coverage Batch: Small Deterministic Packages

- [x] 3.1 Add focused tests for low-risk pure/helper packages with low coverage
- [x] 3.2 Add meaningful assertions for enum, parser, embed, and simple wrapper behavior
- [x] 3.3 Commit the batch separately

## 4. Coverage Batch: CLI/TUI State And Command Surfaces

- [x] 4.1 Add targeted tests for `internal/cli/tuicore/state_update.go`
- [x] 4.2 Add targeted tests for `internal/cli/settings/editor.go` and `internal/cli/settings/menu.go`
- [x] 4.3 Cover command construction, output seams, error paths, and state transitions
- [x] 4.4 Commit the batch separately

## 5. Coverage Batch: Storage And Knowledge

- [x] 5.1 Add focused tests for `internal/storage/facade.go`
- [x] 5.2 Add focused tests for `internal/storagebroker/server.go` and `internal/storagebroker/client.go`
- [x] 5.3 Add focused tests for `internal/knowledge/store.go`
- [x] 5.4 Cover CRUD, error handling, unavailable backends, and boundary conditions
- [x] 5.5 Commit the batch separately

## 6. Coverage Batch: Runtime And Workflow

- [x] 6.1 Add deterministic tests or seams for `internal/adk/agent.go`
- [x] 6.2 Add deterministic tests for `internal/workflow/engine.go`
- [x] 6.3 Add deterministic tests for `internal/turnrunner/runner.go`
- [x] 6.4 Add persistence boundary tests for `internal/turntrace/store.go`
- [x] 6.5 Cover workflow execution/status paths, turn-runner outcomes, and trace persistence boundaries
- [x] 6.6 Commit the batch separately

## 7. Coverage Batch: App Wiring And P2P

- [x] 7.1 Add focused tests for `internal/app/wiring_p2p.go`
- [x] 7.2 Add focused tests for `internal/app/wiring.go`
- [x] 7.3 Add focused tests for `internal/app/tools_meta.go`
- [x] 7.4 Add focused tests for `internal/p2p/handshake/handshake.go`
- [x] 7.5 Commit the batch separately

## 8. Coverage Batch: High-Impact Service Boundaries

- [x] 8.1 Add focused tests for remaining `internal/storagebroker/server.go` dispatch, payload protection, and error paths
- [x] 8.2 Add focused tests for remaining `internal/skill/importer.go` import validation, filesystem, and failure paths
- [x] 8.3 Add focused tests for remaining `internal/receipts/store.go` receipt lifecycle and query boundaries
- [x] 8.4 Commit the batch separately

## 9. Coverage Batch: App And CLI Execution Surfaces

- [x] 9.1 Add focused tests for remaining `internal/app/app.go`, `internal/app/modules.go`, and app wiring branches
- [x] 9.2 Add focused tests for `cmd/lango/main.go` command routing and error exits through seams
- [x] 9.3 Add focused tests for `internal/cli/chat/chat.go`, `internal/cli/settings/editor.go`, `internal/cli/sandbox/sandbox.go`, and `internal/cli/onboard/wizard.go`
- [x] 9.4 Commit the batch separately

## 10. Coverage Batch: Remaining P2P And Payment Surfaces

- [x] 10.1 Add focused tests for remaining `internal/p2p/handshake/handshake.go` stream and approval branches
- [x] 10.2 Add focused tests for high-uncovered P2P protocol, settlement, discovery, and payment paths that do not require external services
- [x] 10.3 Add focused tests for remaining `internal/app/tools_escrow.go` and payment tool wrappers
- [x] 10.4 Commit the batch separately

## 11. Coverage Batch: Cron, Broker, And App Helpers

- [x] 11.1 Add focused tests for `internal/cron/store.go` Ent-backed store and conversion helpers
- [x] 11.2 Add focused tests for remaining `internal/storagebroker/server.go` and `internal/storagebroker/client.go` RPC wrapper branches
- [x] 11.3 Add focused tests for app P2P helper surfaces that do not require external services
- [x] 11.4 Commit the batch separately

## 12. Coverage Batch: Smart Account, Wallet, And Git Bundle

- [x] 12.1 Add focused tests for `internal/wallet/spending.go` limiter behavior and store error paths
- [x] 12.2 Add focused tests for `internal/smartaccount/manager.go` manager helpers and encoding branches
- [x] 12.3 Add focused tests for `internal/p2p/gitbundle/protocol.go` protocol handler dispatch and error branches
- [x] 12.4 Add focused tests for app Smart Account tools and wiring helper branches
- [x] 12.5 Commit the batch separately

## 13. Coverage Batch: Filesystem, Test Utilities, And Knowledge Branches

- [x] 13.1 Add focused tests for `internal/tools/filesystem/filesystem.go` delete, exists, mkdir, copy, and path access branches
- [x] 13.2 Add focused tests for `internal/testutil/mock_session_store.go` CRUD, salt, end, counters, copy semantics, and error injection
- [x] 13.3 Add focused tests for `internal/knowledge/store.go` FTS5 result resolution, learning lookup branches, and low-risk search paths
- [x] 13.4 Commit the batch separately

## 14. Coverage Batch: App P2P, Meta Tools, And Knowledge Store

- [x] 14.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` helper and wiring branches
- [x] 14.2 Add focused tests for deterministic `internal/app/tools_meta.go` tool handlers and helper branches
- [x] 14.3 Add focused tests for remaining low-risk `internal/knowledge/store.go` scoring, stats, external ref, and deletion branches
- [x] 14.4 Run subagent-driven review checkpoints for the batch
- [x] 14.5 Commit the batch separately

## 15. Coverage Batch: CLI Entrypoints, Chat State, Settings Navigation, And Config Redaction

- [x] 15.1 Add focused tests for deterministic `cmd/lango/main.go` helper and adapter branches
- [x] 15.2 Add focused tests for deterministic `internal/cli/chat/chat.go` state, accessor, session-mode, and key-helper branches
- [x] 15.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` redaction, scalar, reflection, formatting, and set-field branches
- [x] 15.4 Add focused tests for deterministic `internal/cli/settings/editor.go` dependency navigation and setup-flow branches
- [x] 15.5 Run subagent-driven review checkpoints for the batch
- [x] 15.6 Commit the batch separately

## 16. Coverage Batch: Gateway, Telegram, And App Adapters

- [x] 16.1 Add focused tests for deterministic `internal/gateway/server.go` setter, status, RPC, and companion branches
- [x] 16.2 Add focused tests for deterministic `internal/channels/telegram/telegram.go` channel identity, send, split, error formatting, and allow-list branches
- [x] 16.3 Add focused tests for deterministic `internal/app/app.go` publish, post-agent wiring, and lifecycle registration branches
- [x] 16.4 Run subagent-driven review checkpoints for the batch
- [x] 16.5 Commit the batch separately

## 17. Coverage Batch: App Wiring, Knowledge Store, And CLI/ADK Surfaces

- [x] 17.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/tools_meta.go`, `internal/app/modules.go`, and related app wiring branches
- [x] 17.2 Add focused tests for deterministic `internal/knowledge/store.go` search, FTS5, payload-protection, deletion, and learning branches
- [x] 17.3 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/chat/chat.go`, and `internal/adk` option/helper branches
- [x] 17.4 Run subagent-driven review checkpoints for the batch
- [x] 17.5 Commit the batch separately

## 18. Coverage Batch: Integrations, Ledgers, And Receipts

- [x] 18.1 Add focused tests for deterministic `internal/channels/discord/discord.go` channel identity, send, split, guild allow-list, and error formatting branches
- [x] 18.2 Add focused tests for deterministic `internal/p2p/gitbundle/bundle.go` diff, apply, hook, and safe-apply branches
- [x] 18.3 Add focused tests for deterministic `internal/runledger/writethrough.go`, `internal/mcp/connection.go`, and `internal/receipts/store.go` branch-heavy helpers
- [x] 18.4 Run subagent-driven review checkpoints for the batch
- [x] 18.5 Commit the batch separately

## 19. Coverage Batch: App, Knowledge, And User-Facing Entrypoints

- [x] 19.1 Add focused tests for deterministic `internal/app` metadata, module, wiring, and app helper branches that do not require live network services
- [x] 19.2 Add focused tests for deterministic `internal/knowledge/store.go` persistence, validation, and branch-heavy query helpers
- [x] 19.3 Add focused tests for deterministic `cmd/lango/main.go` and high-uncovered user-facing CLI helper branches without launching interactive UI loops
- [x] 19.4 Run subagent-driven review checkpoints for the batch
- [x] 19.5 Commit the batch separately

## 20. Coverage Batch: CLI TUI Pages, Settings, And ADK Branches

- [x] 20.1 Add focused tests for deterministic `internal/cli/cockpit/pages/missioncontrol.go` rendering, navigation, filter, and command-creation branches
- [x] 20.2 Add focused tests for deterministic `internal/cli/settings/editor.go` editor state, section navigation, validation, and save/cancel helper branches
- [x] 20.3 Add focused tests for deterministic `internal/adk/agent.go` run helper, collection, diagnostics, and error branch helpers without live provider calls
- [x] 20.4 Run subagent-driven review checkpoints for the batch
- [x] 20.5 Commit the batch separately

## 21. Coverage Batch: Knowledge Atomic Branches And App Wiring Hotspots

- [x] 21.1 Add focused tests for deterministic `internal/knowledge/store.go` atomic payload, FTS5 sync, and error helper branches without relying on generated code coverage
- [x] 21.2 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool handler, dispatcher, and receipt/error branches without live network services
- [x] 21.3 Add focused tests for deterministic `internal/app/modules.go`, `internal/app/app.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` module/wiring helper branches
- [x] 21.4 Run subagent-driven review checkpoints for the batch
- [x] 21.5 Commit the batch separately

## 22. Coverage Batch: CLI Entrypoint, Sandbox Probe, And Skill Importer Branches

- [x] 22.1 Add focused tests for deterministic `cmd/lango/main.go` chat/cockpit/workbench/config command branch helpers without launching interactive UI loops
- [x] 22.2 Add focused tests for deterministic `internal/cli/sandbox/sandbox.go` status collection, test/probe command helpers, and sandbox capability branches without requiring privileged sandbox execution
- [x] 22.3 Add focused tests for deterministic `internal/skill/importer.go` git/http import planning, URL/resource handling, and failure branches without live network access
- [x] 22.4 Run subagent-driven review checkpoints for the batch
- [x] 22.5 Commit the batch separately

## 23. Coverage Batch: Escrow Tools, Economy Wiring, And Chat Approval Branches

- [x] 23.1 Add focused tests for deterministic `internal/app/tools_escrow.go` escrow tool handler success, validation, and error branches without live chain calls
- [x] 23.2 Add focused tests for deterministic `internal/app/wiring_economy.go`, `internal/app/wiring.go`, and related app wiring helper branches without live providers or network services
- [x] 23.3 Add focused tests for deterministic `internal/cli/chat/chat.go` approval/key handling, pending state, and submission guard branches without launching the TUI program
- [x] 23.4 Run subagent-driven review checkpoints for the batch
- [x] 23.5 Commit the batch separately

## 24. Coverage Batch: App Meta, Module Initialization, P2P Wiring, And CLI Entrypoints

- [x] 24.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool composition, dependency-disabled, and handler edge branches without live network services
- [x] 24.2 Add focused tests for deterministic `internal/app/modules.go` module initialization, catalog, status, and disabled-dependency branches without live providers
- [x] 24.3 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, and `cmd/lango/main.go` startup and wiring helper branches without binding external services
- [x] 24.4 Run subagent-driven review checkpoints for the batch
- [x] 24.5 Commit the batch separately

## 25. Coverage Batch: Contract Caller, A2A Card, Cover CLI, And ADK Context Helpers

- [x] 25.1 Add focused tests for deterministic `cmd/lango-cover/main.go` argument, module discovery, threshold, and error branches without external commands
- [x] 25.2 Add focused tests for deterministic `internal/a2a/server.go` agent-card construction, P2P/pricing mutation, route output, and skill extraction branches without network listeners
- [x] 25.3 Add focused tests for deterministic `internal/contract/caller.go` ABI, revert-reason, receipt timeout/context, and local error branches without live RPC services
- [x] 25.4 Add focused tests for deterministic `internal/adk/context_model.go` context item, token estimation, and retrieval merge helper branches without live providers
- [x] 25.5 Run subagent-driven review checkpoints for the batch
- [x] 25.6 Commit the batch separately

## 26. Coverage Batch: Mission Projector, Ontology Service, And App/Main Helper Branches

- [x] 26.1 Add focused tests for deterministic `internal/cli/cockpit/missioncontrol_projector.go` projection helper, status mapping, summary, and enrichment branches without launching the TUI
- [x] 26.2 Add focused tests for deterministic `internal/ontology/service.go` action executor, schema health, type usage, permission, and delegation error branches without external services
- [x] 26.3 Add focused tests for deterministic `internal/app` and `cmd/lango/main.go` helper branches that remain high-uncovered without binding listeners or launching interactive loops
- [x] 26.4 Run subagent-driven review checkpoints for the batch
- [x] 26.5 Commit the batch separately

## 27. Coverage Batch: Meta Tools, Receipts Store, Ontology Tools, And App Wiring Hotspots

- [x] 27.1 Add focused tests for deterministic `internal/app/tools_meta.go` and `internal/app/tools_escrow.go` validation, dispatcher, receipt, and helper branches without live network or chain calls
- [x] 27.2 Add focused tests for deterministic `internal/receipts/store.go` lifecycle, query, filtering, and error branches using in-memory or temp-file state only
- [x] 27.3 Add focused tests for deterministic `internal/ontology/tools.go` tool builder, parameter validation, action/query, and error branches without external services
- [x] 27.4 Add focused tests for deterministic `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` helper branches without binding listeners or starting P2P services
- [x] 27.5 Run subagent-driven review checkpoints for the batch
- [x] 27.6 Commit the batch separately

## 28. Coverage Batch: Gateway, P2P Protocol, Gemini Provider, And Remaining App Entrypoints

- [x] 28.1 Add focused tests for deterministic `internal/gateway/server.go` response handlers, callbacks, router setup, and connection lifecycle branches without binding real listeners
- [x] 28.2 Add focused tests for deterministic `internal/p2p/protocol/remote_agent.go` request/response, timeout, encoding, and error branches with fake streams/hosts only
- [x] 28.3 Add focused tests for deterministic `internal/provider/gemini/gemini.go` provider identity, schema conversion, request assembly, model listing, and error branches without live Gemini calls
- [x] 28.4 Add focused tests for deterministic `internal/app/tools_p2p.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `cmd/lango/main.go` helper branches without starting P2P services or interactive loops
- [x] 28.5 Run subagent-driven review checkpoints for the batch
- [x] 28.6 Commit the batch separately

## 29. Coverage Batch: App Metadata, Entrypoints, Test Utilities, And Economy Hotspots

- [x] 29.1 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` branches that remain high-uncovered without live providers or listeners
- [x] 29.2 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/settings/forms_economy.go`, and `internal/testutil/mock_cron.go` helper branches without launching interactive UI loops or relying on wall-clock scheduling
- [x] 29.3 Add focused tests for deterministic `internal/adk/agent.go`, `internal/bootstrap/phases.go`, and `internal/economy/tools.go` branch-heavy helpers without live model providers, chain RPC, or external services
- [x] 29.4 Run subagent-driven review checkpoints for the batch
- [x] 29.5 Commit the batch separately

## 30. Coverage Batch: Remaining App Hotspots, TUI Form, Smart Account, And Workflow CLI

- [x] 30.1 Add focused tests for remaining deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` branches that still dominate uncovered statement count without live providers, listeners, or chain RPC
- [x] 30.2 Add focused tests for deterministic `internal/cli/tuicore/form.go`, `internal/cli/smartaccount/session.go`, and `internal/cli/workflow/workflow.go` helper branches without launching interactive UI loops
- [x] 30.3 Add focused tests for remaining deterministic `cmd/lango/main.go` and `internal/adk/agent.go` helper branches that are safe to exercise with fake stores, fake agents, and no external services
- [x] 30.4 Run subagent-driven review checkpoints for the batch
- [x] 30.5 Commit the batch separately

## 31. Coverage Batch: Status, Settings Knowledge, P2P Node, And ADK Support Helpers

- [x] 31.1 Add focused tests for deterministic `internal/cli/settings/forms_knowledge.go` form builders and field wiring without launching interactive UI loops
- [x] 31.2 Add focused tests for deterministic `internal/cli/status/status.go` dead-letter bridge, retry/detail, sanitization, and status helper branches using fake stores and command writers
- [x] 31.3 Add focused tests for deterministic `internal/p2p/node.go` constructor/accessor, key path, handler, and peer-found helper branches without external peers
- [x] 31.4 Add focused tests for deterministic `internal/adk` support helpers such as child session service, context helpers, PII adapter, plugin constructors, and tool adapters without live providers
- [x] 31.5 Run subagent-driven review checkpoints for the batch
- [x] 31.6 Commit the batch separately

## 32. Coverage Batch: Automation Settings Forms And Smart Account CLI Branches

- [x] 32.1 Add focused tests for deterministic `internal/cli/settings/forms_automation.go` form builders, field wiring, defaults, and validation branches without launching interactive UI loops
- [x] 32.2 Add focused tests for deterministic `internal/cli/smartaccount/policy.go` and remaining `internal/cli/smartaccount/session.go` command output, cleanup, validation, and error branches without live chain RPC
- [x] 32.3 Run subagent-driven review checkpoints for the batch
- [x] 32.4 Commit the batch separately

## 33. Coverage Batch: Session Store, Storage Broker, App Runtime Helpers

- [x] 33.1 Add focused tests for deterministic `internal/session/ent_store.go` constructor, client/DB accessor, list, timeout annotation, checksum, and migration branches using temp SQLite state only
- [x] 33.2 Add focused tests for deterministic `internal/storagebroker/server.go` dispatch, payload encode/decode, crypto wrapper, and unavailable dependency branches without binding real listeners
- [x] 33.3 Add focused tests for remaining deterministic `internal/app` and `cmd/lango/main.go` runtime helper branches without live providers, external listeners, P2P services, or interactive UI loops
- [x] 33.4 Run subagent-driven review checkpoints for the batch
- [x] 33.5 Commit the batch separately

## 34. Coverage Batch: App Meta, Network, And Wiring Branches

- [x] 34.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool construction and exportability/payment helper branches without live payments, chain RPC, or background workers
- [x] 34.2 Add focused tests for deterministic `internal/app/modules.go` network/intelligence module disabled and lightweight catalog branches without starting P2P, payment, or external services
- [x] 34.3 Add focused tests for deterministic `internal/app/wiring.go` and `internal/app/wiring_p2p.go` helper branches such as security/provenance/agent options and invalid or disabled P2P configuration paths
- [x] 34.4 Run subagent-driven review checkpoints for the batch
- [x] 34.5 Commit the batch separately

## 35. Coverage Batch: Knowledge Wiring, Smart Account Sessions, And ADK Helpers

- [x] 35.1 Add focused tests for deterministic `internal/app/wiring_knowledge.go` knowledge, FTS5 bulk-indexing, skill registry, and adapter helper branches without live providers or external services
- [x] 35.2 Add focused tests for deterministic `internal/cli/smartaccount/session.go` execution helper error paths, cleanup behavior, list status mapping, revoke branches, command validation, and table formatting without live chain RPC
- [x] 35.3 Add focused tests for deterministic `internal/adk` low-risk helper branches that remain high-uncovered without live providers, including context retrieval/formatting and session-service discard/cleanup helpers
- [x] 35.4 Run subagent-driven review checkpoints for the batch
- [x] 35.5 Commit the batch separately

## 36. Coverage Batch: Security Status, P2P Handshake, And CLI Helpers

- [x] 36.1 Add focused tests for deterministic `internal/cli/security/status.go` status helpers, envelope/identity rendering, config fallback, and non-interactive DB-status branches without interactive passphrase prompts
- [x] 36.2 Add focused tests for deterministic `internal/p2p/handshake/handshake.go` stream handler, protocol selection, timestamp, signer, and early failure branches without external peers
- [x] 36.3 Add focused tests for deterministic `cmd/lango/main.go` command startup/config helper branches without launching interactive TUI loops or binding external services
- [x] 36.4 Run subagent-driven review checkpoints for the batch
- [x] 36.5 Commit the batch separately

## 37. Coverage Batch: FTS5, Bootstrap Phases, And App Post-Agent Wiring

- [x] 37.1 Add focused tests for deterministic `internal/search/fts5.go` table lifecycle, insert/update/delete, bulk insert, search, DB accessor, and query sanitization branches using temp SQLite where FTS5 is available, plus fallback SQL-path/error/sanitization coverage when the local SQLite runtime lacks FTS5
- [x] 37.2 Add focused tests for deterministic `internal/bootstrap/phases.go` data-dir, encryption detection, credential acquisition, database-open, security-state, crypto-init, and profile-load branches without interactive prompts
- [x] 37.3 Add focused tests for deterministic `internal/app/app.go` and post-agent P2P wiring branches covering memory/turn callbacks, provenance options, security init, and inbound approval wiring without live providers or external peers
- [x] 37.4 Run subagent-driven review checkpoints for the batch
- [x] 37.5 Commit the batch separately

## 38. Coverage Batch: CLI Settings, Workflow, Config, And App Wiring Hotspots

- [x] 38.1 Add focused tests for deterministic `internal/cli/settings/forms_agent.go` form construction, defaults, validation, dependency wiring, and provider/model field branches without launching interactive UI loops
- [x] 38.2 Add focused tests for deterministic `internal/cli/workflow/workflow.go` command construction, output formatting, validation, and error branches using fake writers and no external services
- [x] 38.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` scalar/reflection formatting, redaction, invalid path, and set-field branches without interactive prompts
- [x] 38.4 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` helper branches that still dominate uncovered statement count without live providers, listeners, or external peers
- [x] 38.5 Run subagent-driven review checkpoints for the batch
- [x] 38.6 Commit the batch separately

## 39. Coverage Batch: Workspace Tools, Run Ledger, And Entrypoint Helpers

- [x] 39.1 Add focused tests for deterministic `internal/app/tools_workspace.go` create/join/leave/list/status helper branches without live P2P peers or external services
- [x] 39.2 Add focused tests for deterministic `internal/runledger/ent_store.go` journal, snapshot, cache, and error branches using isolated test storage
- [x] 39.3 Add focused tests for deterministic `cmd/lango/main.go` chat/cockpit/workbench/config helper branches through existing seams without launching interactive UI loops
- [x] 39.4 Run subagent-driven review checkpoints for the batch
- [x] 39.5 Commit the batch separately

## 40. Coverage Batch: USDC Settlement, App Modules, And Chat State

- [x] 40.1 Add focused tests for deterministic `internal/economy/escrow/usdc_settler.go` address resolution, lock, release/refund, signing, submission retry, and receipt confirmation branches using in-process RPC seams only
- [x] 40.2 Add focused tests for deterministic `internal/app/modules.go` foundation/intelligence/automation/network/extension module branches that do not start live external services
- [x] 40.3 Add focused tests for deterministic `internal/cli/chat/chat.go` state/update/render helper branches without launching Bubble Tea programs
- [x] 40.4 Run subagent-driven review checkpoints for the batch
- [x] 40.5 Commit the batch separately

## 41. Coverage Batch: Test Utilities, Smart Account Policy, And P2P Firewall

- [x] 41.1 Add focused tests for deterministic `internal/testutil/mock_graph.go` CRUD, query, traversal, copy semantics, error injection, counters, clear, and close branches
- [x] 41.2 Add focused tests for deterministic `internal/cli/smartaccount/policy.go` show/set command formatting, output mode validation, seam errors, cleanup, value formatting, and limit parsing branches without live chain RPC
- [x] 41.3 Add focused tests for deterministic `internal/p2p/firewall/firewall.go` ACL validation, rule mutation, rate limiting, reputation outcomes, sanitization, attestation, matching, and copy semantics without live network services
- [x] 41.4 Run subagent-driven review checkpoints for the batch
- [x] 41.5 Commit the batch separately

## 42. Coverage Batch: Receipts, Proactive Librarian, And MCP Connection

- [x] 42.1 Add focused tests for deterministic `internal/receipts/store.go` remaining transaction lifecycle, clone/copy, event trail, validation, and error branches without external services
- [x] 42.2 Add focused tests for deterministic `internal/librarian/proactive_buffer.go` default config, trigger processing, message/observation errors, threshold/cooldown/max-pending behavior, auto-save, dual-save, event publishing, and inquiry creation branches without background sleeps
- [x] 42.3 Add focused tests for deterministic `internal/mcp/connection.go` state/accessor/copy semantics, sandbox decision publishing, disconnect, health-check, timeout, and transport error branches without launching real MCP servers
- [x] 42.4 Run subagent-driven review checkpoints for the batch
- [x] 42.5 Commit the batch separately

## 43. Coverage Batch: App Entrypoints, Channels, Smart Account Wiring, And Turn Runner

- [x] 43.1 Add focused tests for deterministic `internal/app/channels.go` and `internal/app/wiring_smartaccount.go` helper branches without launching external channel services, chain RPC, or background workers
- [x] 43.2 Add focused tests for deterministic `internal/turnrunner/runner.go` lifecycle, event emission, error, cancellation, receipt, and callback branches using fake agents and stores only
- [x] 43.3 Add focused tests for deterministic `cmd/lango/main.go` command startup/config helper branches without launching interactive TUI loops or binding listeners
- [x] 43.4 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, and `internal/app/wiring_p2p.go` branches that remain high-uncovered without live providers, listeners, or external peers
- [x] 43.5 Run subagent-driven review checkpoints for the batch
- [x] 43.6 Commit the batch separately

## 44. Coverage Batch: App Wiring, ADK/Knowledge, Deadletters, And Slack

- [x] 44.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` and related `internal/app/app.go` helper branches without live listeners, peers, or P2P services
- [x] 44.2 Add focused tests for deterministic `internal/app/modules.go` and `internal/app/wiring.go` helper branches without live providers, background workers, or external services
- [x] 44.3 Add focused tests for deterministic `internal/adk/agent.go` and `internal/knowledge/store.go` remaining helper/error branches without live model providers or external services
- [x] 44.4 Add focused tests for deterministic `internal/cli/cockpit/pages/deadletters.go` and `internal/channels/slack/slack.go` branches without launching TUI loops or live Slack clients
- [x] 44.5 Run subagent-driven review checkpoints for the batch
- [x] 44.6 Commit the batch separately

## 45. Coverage Batch: ZKP, Storage Facade, Escrow Monitor, And App Helpers

- [x] 45.1 Add focused tests for deterministic `internal/p2p/zkp/zkp.go` branches without external provers, peers, or network services
- [x] 45.2 Add focused tests for deterministic `internal/storage/facade.go` resolution, fallback, and error branches without external storage services
- [x] 45.3 Add focused tests for deterministic `internal/economy/escrow/hub/monitor.go` event handling, replay, state, and error branches without live chain RPC
- [x] 45.4 Add focused tests for remaining deterministic `internal/app/tools_meta.go`, `internal/app/wiring_p2p.go`, `internal/app/modules.go`, and `internal/app/wiring.go` branches that do not require live providers, listeners, peers, or background workers
- [x] 45.5 Run subagent-driven review checkpoints for the batch
- [x] 45.6 Keep Ent-backed coverage batch 45 test bootstraps race-safe with the shared schema serialization seam and generated enttest template
- [x] 45.7 Commit the batch separately

## 46. Verification And Enforcement

## 46. Coverage Batch: Entrypoints, Payment, Storage Broker, And App Wiring

- [x] 46.1 Add focused tests for deterministic `cmd/lango/main.go` command entrypoint branches that avoid launching interactive TUI loops or real servers
- [x] 46.2 Add focused tests for deterministic `internal/payment/service.go` send, balance, retry, confirmation, and failure-recording branches without live RPC
- [x] 46.3 Add focused tests for deterministic `internal/storagebroker/server.go` request dispatch, config, payment, recall, workflow, and error branches without external brokers
- [x] 46.4 Add focused tests for deterministic `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/tools_meta.go` branches that can be covered without live providers, listeners, peers, or long-running workers
- [x] 46.5 Run subagent-driven review checkpoints for the batch
- [x] 46.6 Commit the batch separately

## 47. Coverage Batch: Gateway Auth, App P2P Tools, ADK Agent, And CLI Workflow

- [x] 47.1 Add focused tests for deterministic `internal/gateway/auth.go` authentication middleware branches without real gateway servers
- [x] 47.2 Add focused tests for deterministic `internal/app/tools_p2p.go` and remaining safe `internal/app/wiring_p2p.go` branches without live peers or listeners
- [x] 47.3 Add focused tests for deterministic `internal/adk/agent.go` execution, callback, and error branches without live providers
- [x] 47.4 Add focused tests for deterministic `internal/cli/workflow/workflow.go` command, validation, and output branches without external services
- [x] 47.5 Run subagent-driven review checkpoints for the batch
- [x] 47.6 Commit the batch separately

## 48. Coverage Batch: Knowledge, App Lifecycle, Provider, And RunLedger

- [x] 48.1 Add focused tests for deterministic `internal/knowledge/store.go` branches using in-memory SQLite/FTS fixtures without external services
- [x] 48.2 Add focused tests for deterministic `internal/app/app.go`, `internal/app/wiring.go`, and remaining safe app lifecycle branches without live listeners, peers, or providers
- [x] 48.3 Add focused tests for deterministic `internal/provider/anthropic/anthropic.go` request building, response mapping, and error handling without live Anthropic API calls
- [x] 48.4 Add focused tests for deterministic `internal/runledger/ent_store.go` and related run ledger storage branches using local Ent test stores
- [x] 48.5 Run subagent-driven review checkpoints for the batch
- [x] 48.6 Commit the batch separately

## 49. Coverage Batch: App Wiring, MCP Manager, And P2P Workspace

- [x] 49.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, and `internal/app/modules.go` branches without live network peers
- [x] 49.2 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/tools_p2p.go`, and safe tool wrapper branches without external RPC calls
- [x] 49.3 Add focused tests for deterministic `internal/mcp/manager.go` lifecycle, registration, and error branches without external MCP servers
- [x] 49.4 Add focused tests for deterministic `internal/p2p/workspace/gossip.go` publish/subscribe/error branches using local fakes
- [x] 49.5 Run subagent-driven review checkpoints for the batch
- [x] 49.6 Commit the batch separately

## 50. Coverage Batch: ADK, Workflow, Contract, Payment, And Tool Surfaces

- [x] 50.1 Add focused tests for deterministic `internal/adk/context_model.go` and `internal/adk/tools.go` helper branches without live model providers
- [x] 50.2 Add focused tests for deterministic `internal/workflow/state.go` and `internal/workflow/tools.go` persistence/query/tool branches using local Ent test stores or fakes only
- [x] 50.3 Add focused tests for deterministic `internal/contract/caller.go` read/write helper branches and revert/receipt handling without live RPC services
- [x] 50.4 Add focused tests for deterministic `internal/tools/payment/payment.go`, `internal/tools/filesystem/tools.go`, `internal/tools/websearch/tools.go`, `internal/tools/webfetch/readability.go`, and `internal/x402` helper branches without external network calls
- [x] 50.5 Add focused tests for deterministic `cmd/lango/main.go`, `cmd/lango/dead_letter_status.go`, and `internal/turntrace/retention.go` helper branches without launching interactive TUI loops or long-running workers
- [x] 50.6 Run subagent-driven review checkpoints for the batch
- [x] 50.7 Commit the batch separately

## 51. Coverage Batch: App Hotspots, Gateway, CLI Workflow, And Bundler

- [x] 51.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, `internal/app/wiring.go`, and `internal/app/app.go` branches that remain high-uncovered without live providers, listeners, peers, or long-running workers
- [x] 51.2 Add focused tests for deterministic `internal/gateway/server.go` chat/RPC/startup branches using local fakes and no externally bound services
- [x] 51.3 Add focused tests for deterministic `internal/smartaccount/bundler/client.go` nonce, gas fee, encoding, response, and error branches using local HTTP/RPC fakes only
- [x] 51.4 Add focused tests for deterministic `cmd/lango/main.go` and `internal/cli/workflow/workflow.go` command/helper branches without launching interactive TUI loops
- [x] 51.5 Add focused tests for remaining deterministic `internal/knowledge/store.go` write, atomic, relevance, and FTS cleanup branches using local SQLite/FTS fixtures only
- [x] 51.6 Run subagent-driven review checkpoints for the batch
- [x] 51.7 Commit the batch separately

## 52. Coverage Batch: App Wiring, CLI Hotspots, Ontology, And X402

- [x] 52.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, and `internal/app/wiring.go` branches that remain top uncovered without live providers, listeners, peers, or long-running workers
- [x] 52.2 Add focused tests for deterministic `internal/app/tools_p2p.go`, `internal/app/wiring_economy.go`, and `internal/app/wiring_knowledge.go` helper/error branches without live P2P peers, chain RPC, or external services
- [x] 52.3 Add focused tests for deterministic `internal/cli/chat/chat.go` and `internal/cli/configcmd/getset.go` helper branches without launching interactive TUI loops
- [x] 52.4 Add focused tests for deterministic `internal/ontology/service.go`, `internal/x402`, and adjacent payment/settlement helper branches without external network or chain services
- [x] 52.5 Run subagent-driven review checkpoints for the batch
- [x] 52.6 Commit the batch separately

## 53. Coverage Batch: App Residuals, Post-Adjudication Status, Browser, Smart Account, And Graph Store

- [x] 53.1 Add focused tests for remaining deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/app.go`, and `internal/app/wiring_economy.go` branches plus the related `internal/economy/negotiation` callback lock-safety branch that can be exercised without live providers, listeners, peers, chain RPC, or long-running workers
- [x] 53.2 Add focused tests for deterministic `internal/postadjudicationstatus/service.go` and `internal/graph/bolt_store.go` status aggregation, receipt, filtering, snapshot, stats, scan, and clear branches using local stores only
- [x] 53.3 Add focused tests for deterministic `internal/tools/browser/browser.go` and `internal/cli/smartaccount/policy.go` session/error/helper branches without launching a real browser or live chain RPC
- [x] 53.4 Run subagent-driven review checkpoints for the batch
- [x] 53.5 Commit the batch separately

## 54. Coverage Batch: App Remaining Hotspots, Run Ledger Validators, P2P Team, Ontology, And Node Helpers

- [x] 54.1 Add focused tests for remaining deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, `internal/app/wiring.go`, and `internal/app/wiring_knowledge.go` helper/error branches that can be exercised without live providers, listeners, peers, chain RPC, or long-running workers
- [x] 54.2 Add focused tests for deterministic `internal/runledger/validators.go` and `internal/p2p/team/tools_escrow.go` validation/tool branches without live chain RPC, P2P peers, or external services
- [x] 54.3 Add focused tests for deterministic `internal/ontology/service.go` and `internal/p2p/node.go` schema, delegation, helper, key-path, peer, and accessor branches without live peers or external services
- [x] 54.4 Run subagent-driven review checkpoints for the batch
- [x] 54.5 Commit the batch separately

## 55. Coverage Batch: App, P2P Settlement, MCP, Discovery, CLI, And Channel Hotspots

- [x] 55.1 Add focused tests for remaining deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, `internal/app/wiring.go`, and `internal/app/tools_p2p.go` helper/error branches that can be exercised without live providers, listeners, peers, chain RPC, or long-running workers
- [x] 55.2 Add focused tests for deterministic `internal/p2p/settlement/service.go`, `internal/mcp/connection.go`, and `internal/p2p/discovery/gossip.go` parsing, validation, nil-dependency, local-store, and fake-host branches without live peers or external network services
- [x] 55.3 Add focused tests for deterministic `internal/cli/chat/chat.go`, `internal/channels/telegram/telegram.go`, and adjacent tool/channel helper branches without launching interactive UI loops or calling live channel APIs
- [x] 55.4 Run subagent-driven review checkpoints for the batch
- [x] 55.5 Commit the batch separately

## 56. Coverage Batch: RAG, Git Bundle, Sandbox, Onboarding, And Residual Hotspots

- [x] 56.1 Add focused tests for deterministic `internal/embedding/rag.go` retrieval, resolver, empty-result, filtering, and error branches using local fakes only
- [x] 56.2 Add focused tests for deterministic `internal/p2p/gitbundle/protocol.go` request parsing, validation, session, size-limit, service-error, and response branches using fake streams/services only
- [x] 56.3 Add focused tests for deterministic `internal/sandbox/docker_runtime.go` request/response assembly and error branches without requiring a live Docker daemon
- [x] 56.4 Add focused tests for deterministic `internal/cli/onboard/wizard.go` state, navigation, validation, rendering, and config update branches without launching interactive TUI loops
- [x] 56.5 Add focused tests for remaining deterministic `internal/app/*`, `internal/p2p/settlement/service.go`, and `internal/mcp/connection.go` helper branches only where they can be covered without live providers, listeners, peers, chain RPC, or external MCP servers
- [x] 56.6 Run subagent-driven review checkpoints for the batch
- [x] 56.7 Commit the batch separately

## 57. Coverage Batch: Remaining High-Uncovered Deterministic Hotspots

- [x] 57.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, `internal/app/wiring.go`, `internal/app/tools_p2p.go`, `internal/app/app.go`, `internal/app/tools_workspace.go`, and `internal/app/wiring_ontology.go` helper/error branches that can be exercised without live providers, listeners, peers, chain RPC, or long-running workers
- [x] 57.2 Add focused tests for deterministic `internal/graph/rag.go`, `internal/p2p/gitbundle/bundle.go`, and `internal/p2p/handshake/handshake.go` branches using local fakes, temp repositories, and in-memory streams only
- [x] 57.3 Add focused tests for deterministic `internal/cli/settings/editor.go`, `internal/cli/smartaccount/paymaster.go`, `internal/p2p/team/tools.go`, and `internal/cron/tools.go` parsing, validation, render, and command-construction branches without launching interactive UI loops, live chain RPC, or external peers
- [x] 57.4 Add focused tests for deterministic `internal/mission/store.go`, `internal/smartaccount/session/manager.go`, `internal/ontology/service.go`, `internal/smartaccount/manager.go`, and `internal/storagebroker/server.go` store/helper/error branches using temp stores and local fakes only
- [x] 57.5 Run subagent-driven review checkpoints for the batch
- [x] 57.6 Commit the batch separately

## 58. Test Naming Cleanup

- [x] 58.1 Rename legacy numbered test-unit files to behavior-oriented names without work-unit numbers
- [x] 58.2 Rename numbered test-unit helper identifiers to behavior-oriented names where they are not fixture values
- [x] 58.3 Run subagent-driven review checkpoints for the naming cleanup
- [x] 58.4 Commit the naming cleanup separately

## 59. Verification And Enforcement

- [ ] 59.1 Run `go build ./...`
- [ ] 59.2 Run `go test ./...`
- [ ] 59.3 Run the non-generated coverage report and confirm coverage is at least 90%
- [ ] 59.4 Run the executable 90% coverage gate and confirm it passes
- [ ] 59.5 Run `git diff --check`
- [ ] 59.6 Run subagent-driven spec and code-quality review checkpoints
- [ ] 59.7 Run `openspec validate --all --strict`
- [ ] 59.8 Archive the completed OpenSpec change

## 60. Coverage Batch: Residual Deterministic Hotspots

- [x] 60.1 Add focused tests for remaining deterministic `internal/app` provenance and channel sender helpers without live channel or P2P services
- [x] 60.2 Add focused tests for deterministic extension registry, gateway auth, and security CLI helper branches using local fakes and temp files only
- [x] 60.3 Run subagent-driven review checkpoints for the batch
- [x] 60.4 Commit the batch separately

## 61. Coverage Batch: App Module And Wiring Residuals

- [x] 61.1 Add focused tests for deterministic `internal/app/modules.go` branches without provider credentials, listeners, or live P2P services
- [x] 61.2 Add focused tests for deterministic `internal/app/wiring.go` and `internal/app/wiring_provenance.go` helper and early-return branches
- [x] 61.3 Run subagent-driven review checkpoints for the batch
- [x] 61.4 Commit the batch separately

## 62. Coverage Batch: App Meta Tool Early Handlers

- [x] 62.1 Add focused tests for deterministic `internal/app/tools_meta.go` knowledge, learning, skill, exportability, and dry-run cleanup count branches without live network services
- [x] 62.2 Run subagent-driven review checkpoints for the batch
- [x] 62.3 Commit the batch separately

## 63. Coverage Batch: App Wiring And P2P Residuals

- [x] 63.1 Add focused tests for deterministic `internal/app/wiring.go` session-store fallback error branches without live providers
- [x] 63.2 Add focused tests for deterministic `internal/app/wiring_p2p.go` branches using only local ephemeral P2P resources and cleanup
- [x] 63.3 Run subagent-driven review checkpoints for the batch
- [x] 63.4 Commit the batch separately

## 64. Coverage Batch: Browser Tool And Config Save Residuals

- [x] 64.1 Add focused tests for deterministic `internal/tools/browser/tools.go` handler validation, P2P pre-session blocking, and search/extract handler behavior without launching a live browser
- [x] 64.2 Add focused tests for deterministic `internal/cli/configcmd/getset.go` config-save failure handling and cleanup behavior
- [x] 64.3 Run subagent-driven review checkpoints for the batch
- [x] 64.4 Commit the batch separately

## 65. Naming Cleanup: Replace Legacy Sequence Labels

- [x] 65.1 Rename legacy sequential internal planning and archived OpenSpec artifacts to `slice-*`
- [x] 65.2 Replace old sequential wording in affected internal planning, archived OpenSpec, and public spec references with `Slice N`/`slice`
- [x] 65.3 Run subagent-driven review checkpoints for the naming cleanup
- [x] 65.4 Commit the naming cleanup separately

## 66. Coverage Batch: App Module Residuals

- [x] 66.1 Add focused tests for deterministic `internal/app/modules.go` foundation module session-store error wrapping without live providers
- [x] 66.2 Add focused tests for deterministic app module lifecycle component registration without starting long-running workers
- [x] 66.3 Run subagent-driven review checkpoints for the batch
- [x] 66.4 Commit the batch separately

## 67. Coverage Batch: Run Ledger Write-Through Residuals

- [x] 67.1 Add focused tests for deterministic workflow write-through shadow, degraded sync, status mapping, and drift branches using in-memory stores only
- [x] 67.2 Add focused tests for deterministic background write-through disabled and error branches using local fakes only
- [x] 67.3 Run subagent-driven review checkpoints for the batch
- [x] 67.4 Commit the batch separately

## 68. Coverage Batch: Knowledge Wiring FTS5 Residuals

- [x] 68.1 Add focused tests for deterministic `internal/app/wiring_knowledge.go` FTS5 table creation and bulk-insert error branches using temp SQLite only
- [x] 68.2 Run subagent-driven review checkpoints for the batch
- [x] 68.3 Commit the batch separately

## 69. Coverage Batch: Smart Account Policy Helper Residuals

- [x] 69.1 Add a minimal test seam for deterministic `internal/cli/smartaccount/policy.go` helper branches without live RPC, wallet, or bootstrap services
- [x] 69.2 Add focused tests for policy show/set helper success, error, and policy mutation branches using local fakes only
- [x] 69.3 Run subagent-driven review checkpoints for the batch
- [x] 69.4 Commit the batch separately

## 70. Coverage Batch: Receipt Settlement Evidence Residuals

- [x] 70.1 Add focused tests for deterministic `internal/receipts/store.go` escrow refund failure, dispute hold success, and adjudication failure guard branches using in-memory stores only
- [x] 70.2 Add focused tests for remaining escrow adjudication validation branches without external services
- [x] 70.3 Run subagent-driven review checkpoints for the batch
- [x] 70.4 Commit the batch separately

## 71. Coverage Batch: Config Command And Workbench Helper Residuals

- [x] 71.1 Add focused tests for deterministic `internal/cli/configcmd/getset.go` redaction, path, suggestion, collection, and print error residual branches
- [x] 71.2 Add focused tests for deterministic `internal/cli/cockpit/pages/missioncontrol_workbench.go` starter prompt, follow-up, hint, and footer residual branches without launching TUI loops
- [x] 71.3 Run subagent-driven development and review checkpoints for the batch
- [x] 71.4 Commit the batch separately

## 72. Coverage Batch: Run Ledger Ent Store And ZKP Residuals

- [x] 72.1 Add focused tests for deterministic `internal/runledger/ent_store.go` cached snapshot, journal tail, prune, and retry residual branches using local SQLite Ent stores only
- [x] 72.2 Add focused tests for deterministic `internal/p2p/zkp/zkp.go` Groth16 export and SRS file residual branches without network services
- [x] 72.3 Run subagent-driven development and review checkpoints for the batch
- [x] 72.4 Commit the batch separately

## 73. Coverage Batch: Security Migration And Filesystem Residuals

- [x] 73.1 Add focused tests for deterministic `internal/security/migrate_envelope.go` checksum-optional migration and quoted backup path residual branches
- [x] 73.2 Add focused tests for deterministic `internal/tools/filesystem/filesystem.go` read, edit, write, metadata, and validation residual branches using temp dirs only
- [x] 73.3 Run subagent-driven development and review checkpoints for the batch
- [x] 73.4 Commit the batch separately

## 74. Coverage Batch: ADK Context Model And Knowledge Retriever Residuals

- [x] 74.1 Add focused tests for deterministic `internal/adk/context_model.go` compaction sync, recall, catalog, and mode prompt branches using local fakes only
- [x] 74.2 Add focused tests for deterministic `internal/knowledge/retriever.go` skill, inquiry, default limit, unsupported layer, and keyword sanitization branches using local fakes only
- [x] 74.3 Run subagent-driven development and review checkpoints for the batch
- [x] 74.4 Prepare the batch for a separate scoped commit

## 75. Coverage Batch: Run Ledger Store And Cockpit Projector Residuals

- [x] 75.1 Add focused tests for deterministic `internal/runledger/ent_store.go` query, snapshot tail, validation rollback, prune, and retry residual branches using temp SQLite Ent stores only
- [x] 75.2 Add focused tests for deterministic `internal/cli/cockpit/missioncontrol_projector.go` status, time, and active-agent summary residual branches using local data only
- [x] 75.3 Run subagent-driven development and review checkpoints for the batch
- [x] 75.4 Prepare the batch for a separate scoped commit

## 76. Coverage Batch: Security Migration Residuals

- [x] 76.1 Add focused tests for deterministic `internal/security/migrate_envelope.go` envelope persistence, backup, decrypt, rekey, transaction, and invalid master-key residual branches using temp SQLite only
- [x] 76.2 Run subagent-driven development and review checkpoints for the batch
- [x] 76.3 Prepare the batch for a separate scoped commit

## 77. Coverage Batch: Economy Wiring Residuals

- [x] 77.1 Add focused tests for deterministic `internal/app/wiring_economy.go` pricing, paygate wiring, negotiation expiry, P2P negotiator, and escrow config residual branches without live network or RPC services
- [x] 77.2 Run subagent-driven development and review checkpoints for the batch
- [x] 77.3 Prepare the batch for a separate scoped commit

## 78. Coverage Batch: CLI HTTP And Cockpit Activity Helpers

- [x] 78.1 Add focused tests for deterministic `internal/cli/clihttp` POST, output, JSON printing, and gateway error branches using local `httptest` servers only
- [x] 78.2 Add focused tests for deterministic `internal/cli/cockpit` mission activity reset, assistant fallback, and activity constructor branches without interactive TUI loops
- [x] 78.3 Run subagent-driven development and review checkpoints for the batch
- [x] 78.4 Prepare the batch for a separate scoped commit

## 79. Coverage Batch: P2P Protocol Team And Negotiation Residuals

- [x] 79.1 Add focused tests for deterministic `internal/p2p/protocol` team router typed dispatch, handler-missing, payload marshal/decode, and unknown request branches
- [x] 79.2 Add focused tests for deterministic negotiation handler nil, success, backend error, and tolerant payload decode branches without libp2p streams
- [x] 79.3 Run subagent-driven development and review checkpoints for the batch
- [x] 79.4 Prepare the batch for a separate scoped commit

## 80. Coverage Batch: Network Module And Lifecycle Residuals

- [x] 80.1 Add focused tests for deterministic `internal/app/modules.go` no-payment economy escrow and sentinel catalog/tool registration branches without RPC or P2P startup
- [x] 80.2 Add focused tests for deterministic `internal/app/app.go` channel start error lifecycle branch using a fake channel without binding a gateway listener
- [x] 80.3 Run subagent-driven development and review checkpoints for the batch
- [x] 80.4 Prepare the batch for a separate scoped commit

## 81. Coverage Batch: App Helper Residuals

- [x] 81.1 Add focused tests for deterministic `internal/app/compaction_sync_holder.go` waiter swap, delegation, and nil reset branches
- [x] 81.2 Add focused tests for deterministic `internal/app/bridge_team_metrics.go` counters and `internal/app/bridge_workspace_team.go` ID truncation helpers using in-memory event bus only
- [x] 81.3 Run subagent-driven development and review checkpoints for the batch
- [x] 81.4 Prepare the batch for a separate scoped commit

## 82. Coverage Batch: App Wiring Prompt, Catalog, And Auth Residuals

- [x] 82.1 Add focused tests for deterministic `internal/app/wiring.go` default prompt builder and custom sub-agent prompt directory fallback branches
- [x] 82.2 Add focused tests for deterministic `internal/app/wiring.go` catalog mode empty allowlist fallback and disabled category reporting branches
- [x] 82.3 Add focused tests for deterministic `internal/app/wiring.go` auth manager success wiring using a local OIDC discovery server only
- [x] 82.4 Run subagent-driven development and review checkpoints for the batch
- [x] 82.5 Prepare the batch for a separate scoped commit

## 83. Coverage Batch: Cockpit Mission Control Projector Residuals

- [x] 83.1 Add focused tests for deterministic `internal/cli/cockpit/missioncontrol_projector.go` nil receiver and empty activity projection branches
- [x] 83.2 Add focused tests for deterministic learning suggestion fallback title and collaboration degradation branches
- [x] 83.3 Run subagent-driven development and review checkpoints for the batch
- [x] 83.4 Prepare the batch for a separate scoped commit

## 84. Coverage Batch: Bootstrap Confirmation Residuals

- [x] 84.1 Add a focused test for deterministic `internal/bootstrap/phases.go` confirmation prompt writer error propagation
- [x] 84.2 Run subagent-driven development and review checkpoints for the batch
- [x] 84.3 Prepare the batch for a separate scoped commit

## 85. Coverage Batch: App P2P Sandbox Wiring Residuals

- [x] 85.1 Add focused tests for deterministic `internal/app/app.go` P2P tool isolation subprocess and native-container sandbox executor wiring branches
- [x] 85.2 Add focused tests for deterministic required-container fail-closed sandbox wiring branch without executing sandbox commands
- [x] 85.3 Run subagent-driven development and review checkpoints for the batch
- [x] 85.4 Prepare the batch for a separate scoped commit

## 86. Coverage Batch: P2P Workspace CLI Residuals

- [x] 86.1 Add focused tests for deterministic `internal/cli/p2p/workspace.go` table output branches without network startup
- [x] 86.2 Add focused tests for deterministic workspace manager missing-bootstrap-config and nil-member projection branches
- [x] 86.3 Run subagent-driven development and review checkpoints for the batch
- [x] 86.4 Prepare the batch for a separate scoped commit

## 87. Coverage Batch: Session Ent Store History Residuals

- [x] 87.1 Add focused tests for deterministic `internal/session/ent_store.go` initial history persistence with author and plaintext tool call projection
- [x] 87.2 Add focused tests for deterministic session history/tool call round-trip without payload protector encryption
- [x] 87.3 Run subagent-driven development and review checkpoints for the batch
- [x] 87.4 Prepare the batch for a separate scoped commit

## 88. Coverage Batch: P2P Identity Bundle Provider Residuals

- [x] 88.1 Add focused tests for deterministic `internal/p2p/identity/bundle_provider.go` public key, DID verification, and missing legacy provider branches
- [x] 88.2 Add focused tests for deterministic post-quantum signing key absent/present accessor and signing branches
- [x] 88.3 Run subagent-driven development and review checkpoints for the batch
- [x] 88.4 Prepare the batch for a separate scoped commit
