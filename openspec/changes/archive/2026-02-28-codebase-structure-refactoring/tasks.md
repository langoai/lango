## 1. Phase 1A — Split internal/app/tools.go

- [x] 1.1 Create `tools_exec.go` with `buildExecTools`
- [x] 1.2 Create `tools_filesystem.go` with `buildFilesystemTools`
- [x] 1.3 Create `tools_browser.go` with `buildBrowserTools`
- [x] 1.4 Create `tools_meta.go` with `buildMetaTools`
- [x] 1.5 Create `tools_security.go` with `buildCryptoTools`, `buildSecretsTools`
- [x] 1.6 Create `tools_automation.go` with `buildCronTools`, `buildBackgroundTools`, `buildWorkflowTools`
- [x] 1.7 Create `tools_p2p.go` with `buildP2PTools`, `buildP2PPaymentTool`
- [x] 1.8 Create `tools_data.go` with `buildGraphTools`, `buildRAGTools`, `buildMemoryAgentTools`, `buildPaymentTools`, `buildLibrarianTools`
- [x] 1.9 Remove moved functions from `tools.go` and clean up imports
- [x] 1.10 Verify `go build ./internal/app/...` passes

## 2. Phase 1B — Split internal/app/wiring.go

- [x] 2.1 Create `wiring_knowledge.go` with `knowledgeComponents`, `initKnowledge`, `initSkills`, `initConversationAnalysis`, adapter types
- [x] 2.2 Create `wiring_memory.go` with `memoryComponents`, `initMemory`
- [x] 2.3 Create `wiring_embedding.go` with `embeddingComponents`, `initEmbedding`
- [x] 2.4 Create `wiring_graph.go` with `graphComponents`, `initGraphStore`, `wireGraphCallbacks`, `initGraphRAG`, `ragServiceAdapter`
- [x] 2.5 Create `wiring_payment.go` with `paymentComponents`, `initPayment`, `x402Components`, `initX402`
- [x] 2.6 Create `wiring_p2p.go` with `p2pComponents`, `initP2P`, `payGateAdapter`, `initZKP`
- [x] 2.7 Create `wiring_automation.go` with `agentRunnerAdapter`, `initCron`, `initBackground`, `initWorkflow`
- [x] 2.8 Create `wiring_librarian.go` with `librarianComponents`, `initLibrarian`
- [x] 2.9 Remove moved functions/types from `wiring.go` and clean up imports
- [x] 2.10 Verify `go build ./internal/app/...` passes

## 3. Phase 1C — Split internal/cli/settings/forms_impl.go

- [x] 3.1 Create `forms_knowledge.go` with `NewKnowledgeForm`, `NewSkillForm`, `NewObservationalMemoryForm`, `NewEmbeddingForm`, `NewGraphForm`, `NewLibrarianForm`
- [x] 3.2 Create `forms_automation.go` with `NewCronForm`, `NewBackgroundForm`, `NewWorkflowForm`
- [x] 3.3 Create `forms_security.go` with `NewSecurityForm`, `NewDBEncryptionForm`, `NewKMSForm`
- [x] 3.4 Create `forms_p2p.go` with `NewP2PForm`, `NewP2PZKPForm`, `NewP2PPricingForm`, `NewP2POwnerProtectionForm`, `NewP2PSandboxForm`
- [x] 3.5 Create `forms_agent.go` with `NewMultiAgentForm`, `NewA2AForm`, `NewPaymentForm`
- [x] 3.6 Remove moved functions from `forms_impl.go` and clean up imports
- [x] 3.7 Verify `go build ./internal/cli/settings/...` passes

## 4. Phase 2A — Split internal/config/types.go

- [x] 4.1 Create `types_security.go` with security/auth domain types and methods
- [x] 4.2 Create `types_knowledge.go` with knowledge/embedding/RAG domain types and migration functions
- [x] 4.3 Create `types_p2p.go` with P2P/ZKP/firewall domain types
- [x] 4.4 Create `types_automation.go` with cron/background/workflow/payment/A2A domain types
- [x] 4.5 Remove moved types/functions from `types.go` and clean up imports
- [x] 4.6 Verify `go build ./internal/config/...` passes

## 5. Final Verification

- [x] 5.1 Run `go build ./...` — full project build passes
- [x] 5.2 Run `go test ./...` — all tests pass
