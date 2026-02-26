## 1. Phase 1: P2P Core Package Tests

- [x] 1.1 Create `internal/p2p/discovery/gossip_test.go` — KnownPeers, FindByCapability, FindByDID, RevokeDID, IsRevoked, SetMaxCredentialAge.
- [x] 1.2 Create `internal/p2p/discovery/agentad_test.go` — AdService creation, StoreAd, Discover, DiscoverByCapability, matchesTags, ZK credential verification, timestamp ordering.
- [x] 1.3 Create `internal/p2p/identity/identity_test.go` — DIDFromPublicKey, ParseDID/VerifyDID roundtrip, WalletDIDProvider caching, wallet error handling.

## 2. Phase 2: CLI Tests

- [x] 2.1 Create `internal/cli/p2p/p2p_test.go` — Command tree: 11 subcommands (status, peers, connect, disconnect, firewall, discover, identity, reputation, pricing, session, sandbox), sub-subcommands, --json flag.
- [x] 2.2 Create `internal/cli/security/security_test.go` — Command tree: 7 subcommands (migrate-passphrase, secrets, status, keyring, db-migrate, db-decrypt, kms), boolToStatus, isKMSProvider utility tests.

## 3. Phase 3: Infrastructure Package Tests

- [x] 3.1 Create `internal/workflow/dag_test.go` — Linear/diamond/parallel DAG, circular dependency detection, TopologicalSort, Roots, Ready with completion states.
- [x] 3.2 Create `internal/workflow/parser_test.go` — YAML parsing, Validate (empty name, no steps, empty step ID, duplicate IDs, unknown dependency, agents, circular deps).
- [x] 3.3 Create `internal/workflow/template_test.go` — RenderPrompt (no placeholders, substitution, missing key, hyphenated/underscored IDs), placeholderRe regex.
- [x] 3.4 Create `internal/background/manager_test.go` — Manager defaults, custom values, Submit+List, max tasks, Cancel/Status/Result not found, runner error, Status enum.

## 4. Phase 4: Security/Sandbox Tests

- [x] 4.1 Create `internal/security/kms_factory_test.go` — KMSProviderName.Valid for 4 providers + invalid, constants, NewKMSProvider unknown provider error.
- [x] 4.2 Create `internal/security/kms_checker_test.go` — KMSHealthChecker default/custom probe interval, healthy/unhealthy encrypt+decrypt, cache fresh/expired.
- [x] 4.3 Create `internal/sandbox/subprocess_test.go` — NewSubprocessExecutor, cleanEnv, workerFlag, IsWorkerMode default.

## 5. Phase 5: Additional Package Tests

- [x] 5.1 Create `internal/librarian/inquiry_processor_test.go` — stripCodeFence, parseAnswerMatches, parseAnalysisOutput, buildMatchPrompt, NewInquiryProcessor, confidence filtering.
- [x] 5.2 Create `internal/cli/payment/payment_test.go` — Command tree: 5 subcommands (balance, history, limits, info, send), --json flag, --force flag, required flags.
- [x] 5.3 Create `internal/app/p2p_routes_test.go` — p2pPricingHandler (all prices, specific tool, unknown fallback, disabled), p2pReputationHandler (missing peer_did, nil reputation).

## 6. Phase 6: Documentation Fixes

- [x] 6.1 Update `docs/configuration.md` — Add `p2p.requireSignedChallenge` to P2P table.
- [x] 6.2 Update `docs/configuration.md` — Add `p2p.zkp.srsMode`, `p2p.zkp.srsPath`, `p2p.zkp.maxCredentialAge` to ZKP keys.
- [x] 6.3 Update `docs/configuration.md` — Add P2P Tool Isolation section with all `p2p.toolIsolation.*` and `p2p.toolIsolation.container.*` keys.
- [x] 6.4 Update `docs/configuration.md` — Update JSON example to include all missing keys.

## 7. Verification

- [ ] 7.1 Run `go vet ./...` — static analysis passes (blocked by Go 1.25.4 toolchain in sandbox).
- [ ] 7.2 Run `go test ./...` — all tests pass (blocked by Go 1.25.4 toolchain in sandbox).
