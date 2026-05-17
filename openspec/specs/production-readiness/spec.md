## Purpose

Capability spec for production-readiness. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Unsupported security provider produces actionable error
The system SHALL reject unsupported security provider names at config-time with an error message listing all valid provider options (local, rpc, aws-kms, gcp-kms, azure-kv, pkcs11).

#### Scenario: Enclave provider is rejected as invalid config
- **WHEN** security.signer.provider is set to `enclave`
- **THEN** config validation rejects the value as invalid
- **AND** the validation error lists only `local`, `rpc`, `aws-kms`, `gcp-kms`, `azure-kv`, and `pkcs11` as valid providers

### Requirement: Telegram media download completes successfully
The system SHALL download file content from Telegram's file API via HTTP GET with a 30-second timeout and return the raw bytes.

#### Scenario: File body read failure preserves the I/O cause
- **WHEN** the Telegram file API returns a 200 response but reading the body fails
- **THEN** the system returns an error identifying the body-read failure
- **AND** SHALL preserve the underlying I/O cause instead of collapsing into the empty-body path

### Requirement: No dead code or context.TODO in x402 package
The x402 package SHALL contain no unused exported functions and no `context.TODO()` calls.

#### Scenario: NewX402Client removed
- **WHEN** the codebase is scanned for calls to NewX402Client
- **THEN** no references exist and the function is not present in the source

#### Scenario: No context.TODO remaining
- **WHEN** the x402 package is scanned for context.TODO()
- **THEN** zero occurrences are found

### Requirement: GVisor stub behavior is documented and tested
The GVisor runtime stub SHALL clearly document its stub nature and have tests verifying stub behavior.

#### Scenario: Explicit gVisor runtime request returns actionable unavailable error
- **WHEN** the container executor is configured with runtime `gvisor`
- **AND** the current build still uses the gVisor stub
- **THEN** executor construction SHALL fail with an error that explains the requested gVisor runtime is unavailable

### Requirement: Container runtime fail-closed errors stay actionable
Sandbox runtime selection SHALL keep operator-facing unavailable errors specific enough to explain which runtime or policy path failed.

#### Scenario: Explicit Docker runtime request names the unavailable runtime
- **WHEN** the container executor is configured with runtime `docker`
- **AND** Docker is unavailable in the current environment
- **THEN** executor construction SHALL fail with an error that explains the requested Docker runtime is unavailable

#### Scenario: Require-container fail-closed path names the container requirement
- **WHEN** the container executor is configured with `RequireContainer=true`
- **AND** no container runtime is available
- **THEN** executor construction SHALL fail with an error that explains container runtime availability is required rather than silently falling back to native execution

### Requirement: Wallet package has unit test coverage
The wallet package SHALL have tests covering address derivation, transaction signing, message signing, composite fallback logic, wallet creation, and RPC dispatching.

#### Scenario: Local wallet signs transaction
- **WHEN** SignTransaction is called with a valid key in SecretsStore
- **THEN** the signature is valid and the public key can be recovered

#### Scenario: Composite wallet falls back on primary failure
- **WHEN** the primary wallet provider is disconnected
- **THEN** the composite wallet delegates to the fallback provider

#### Scenario: Wallet creation stores recoverable key
- **WHEN** CreateWallet is called
- **THEN** the stored key can be retrieved and derives the same address

### Requirement: Security KeyRegistry and SecretsStore have unit test coverage
The security package SHALL have tests covering full CRUD operations on KeyRegistry and SecretsStore with mock CryptoProvider.

#### Scenario: KeyRegistry register and retrieve
- **WHEN** a key is registered via RegisterKey
- **THEN** GetKey returns the same key with correct metadata

#### Scenario: SecretsStore encrypt and decrypt roundtrip
- **WHEN** a secret is stored via Store
- **THEN** Get returns the decrypted original value

#### Scenario: SecretsStore with no encryption key
- **WHEN** Store is called with no encryption key registered
- **THEN** it returns ErrNoEncryptionKeys

### Requirement: Payment service has unit test coverage

The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Missing transaction RPC fails closed during submission
- **WHEN** `payment.Service.submitWithRetry` runs without a configured transaction RPC client
- **THEN** the returned error SHALL preserve an actionable transaction-RPC-unavailable cause instead of panicking

#### Scenario: Missing transaction RPC fails closed during confirmation
- **WHEN** `payment.Service.waitForConfirmation` runs without a configured transaction RPC client
- **THEN** the returned error SHALL preserve an actionable receipt-RPC-unavailable cause instead of panicking

### Requirement: Smart account packages have unit test coverage
The smartaccount package SHALL have tests covering factory CREATE2 computation, session key crypto, ABI encoding, paymaster errors, policy syncing, and type methods.

#### Scenario: CREATE2 address is deterministic
- **WHEN** ComputeAddress is called with identical inputs
- **THEN** it produces the same address every time

#### Scenario: Session key serialize/deserialize roundtrip
- **WHEN** a session key is serialized then deserialized
- **THEN** the restored key equals the original

#### Scenario: Policy drift detection
- **WHEN** DetectDrift is called with matching on-chain and Go-side policies
- **THEN** no drift is reported

### Requirement: Economy risk package has unit test coverage
The economy/risk package SHALL have tests covering risk factor computation and strategy selection matrix.

#### Scenario: Risk classification boundaries
- **WHEN** computeRiskScore produces boundary values
- **THEN** classifyRisk returns the correct risk level at each threshold

#### Scenario: Strategy matrix covers all combinations
- **WHEN** SelectStrategy is called with all 9 trust/verifiability combinations
- **THEN** each combination returns the expected strategy

### Requirement: P2P team conflict resolution has unit test coverage
The p2p/team package SHALL have tests covering all 4 conflict resolution strategies.

#### Scenario: TrustWeighted picks fastest successful agent
- **WHEN** ResolveConflict is called with TrustWeighted strategy
- **THEN** the fastest successful agent's result is selected

#### Scenario: FailOnConflict rejects disagreement
- **WHEN** ResolveConflict is called with FailOnConflict and conflicting results
- **THEN** an error is returned

### Requirement: P2P protocol messages and remote agent have unit test coverage
The p2p/protocol package SHALL have tests covering ResponseStatus validation, RequestType constants, and RemoteAgent accessors.

#### Scenario: ResponseStatus.Valid for all statuses
- **WHEN** Valid() is called on each defined ResponseStatus
- **THEN** it returns true for valid statuses and false for invalid ones

#### Scenario: RemoteAgent field population
- **WHEN** NewRemoteAgent is called with a config
- **THEN** all accessor methods return the configured values

### Requirement: x402 static quality gates enforce dead-code and TODO hygiene

The `x402` package SHALL have executable tests that prevent `context.TODO()` reintroduction and legacy `NewX402Client` references.

#### Scenario: x402 package source contains no context.TODO calls
- **WHEN** the `x402` package quality guard tests scan `internal/x402` source files
- **THEN** the scan SHALL find zero `context.TODO()` occurrences

#### Scenario: Repository contains no legacy x402 client factory references
- **WHEN** the `x402` package quality guard tests scan repository Go files
- **THEN** the scan SHALL find zero `NewX402Client` references

### Requirement: Telegram download request shape is verified

Telegram media download coverage SHALL verify not only success and failure outcomes, but also the outgoing HTTP request contract.

#### Scenario: Telegram download uses HTTP GET with a timeout-backed context
- **WHEN** the Telegram file download regression exercises `DownloadFile`
- **THEN** the outgoing request SHALL use the HTTP GET method
- **AND** the request context SHALL carry a deadline derived from the 30-second download timeout contract

### Requirement: Payment history ordering is directly verified

Payment history coverage SHALL verify that the newest transaction is returned first and that history limits apply on top of that descending ordering.

#### Scenario: Payment history returns newest record first
- **WHEN** payment history contains multiple records with distinct `created_at` values
- **THEN** the history response SHALL return them in descending `created_at` order

#### Scenario: Payment history limit keeps the newest record
- **WHEN** payment history is queried with `limit=1`
- **THEN** the response SHALL contain the newest record rather than an arbitrary earlier record

### Requirement: RPC wallet cleans pending request state on non-success paths

RPC wallet coverage SHALL verify that pending request bookkeeping is cleaned up not only after successful responses, but also after sender errors and cancelled contexts.

#### Scenario: Address sender error leaves no pending request entry
- **WHEN** `RPCWallet.Address` fails because the sender returns an error
- **THEN** the pending address request map SHALL be empty after the call returns

#### Scenario: SignMessage sender error leaves no pending request entry
- **WHEN** `RPCWallet.SignMessage` fails because the sender returns an error
- **THEN** the pending sign-message request map SHALL be empty after the call returns

#### Scenario: Address context cancellation leaves no pending request entry
- **WHEN** `RPCWallet.Address` exits because the context is cancelled
- **THEN** the pending address request map SHALL be empty after the call returns

### Requirement: RPC wallet pending-state cleanup is symmetric across request kinds

RPC wallet coverage SHALL verify that pending request bookkeeping is cleaned up consistently for both transaction-signing and message-signing request lifecycles.

#### Scenario: SignTransaction cleanup holds across response and error paths
- **WHEN** `RPCWallet.SignTransaction` exits through a response, companion error, sender error, timeout, or cancelled context
- **THEN** the pending sign-transaction request map SHALL be empty after the call returns

#### Scenario: SignMessage cleanup holds across response and error paths
- **WHEN** `RPCWallet.SignMessage` exits through a response, companion error, sender error, timeout, or cancelled context
- **THEN** the pending sign-message request map SHALL be empty after the call returns

### Requirement: RPC wallet address cleanup holds across timeout and companion-error exits

RPC wallet address lifecycle coverage SHALL explicitly verify timeout and companion-error teardown in addition to the previously covered success, sender-error, and cancellation paths.

#### Scenario: Address timeout clears pending state
- **WHEN** `RPCWallet.Address` exits because the request times out
- **THEN** the pending address request map SHALL be empty after the call returns

#### Scenario: Address companion error clears pending state
- **WHEN** `RPCWallet.Address` exits because the companion responds with an error
- **THEN** the pending address request map SHALL be empty after the call returns

### Requirement: x402 legacy client factory definition stays absent

The `x402` package SHALL have an executable test that prevents the legacy `NewX402Client` function definition from reappearing in package source.

#### Scenario: x402 package source contains no legacy client factory definition
- **WHEN** the x402 quality guard tests scan `internal/x402` source files
- **THEN** the scan SHALL find zero `NewX402Client` function definitions

### Requirement: Mission and proposal service dependency guards stay actionable

Core mission/proposal services SHALL preserve actionable fail-closed errors when required backing dependencies are missing.

#### Scenario: Missing mission store fails closed across mission lifecycle entrypoints
- **WHEN** mission lifecycle methods such as `StartMission`, `AcceptProposal`, or `AttachExecution` run without a configured mission store
- **THEN** each call SHALL return an error that identifies the attempted mission action
- **AND** SHALL preserve the actionable `mission store is required` cause instead of panicking

#### Scenario: Missing proposal registry fails closed across proposal entrypoints
- **WHEN** proposal service methods such as `UpsertLearningSuggestion`, `Accept`, or `PruneExpired` run without a configured proposal registry
- **THEN** each call SHALL return an error that identifies the attempted proposal action
- **AND** SHALL preserve the actionable `proposal registry is required` cause instead of panicking

#### Scenario: Missing proposal preparer fails closed during learning-suggestion preparation
- **WHEN** `proposal.Service.UpsertLearningSuggestion` runs with no configured proposal preparer
- **THEN** the call SHALL return an actionable `proposal preparer is required` error instead of panicking

### Requirement: Payment gate dependency guards stay actionable

The direct-payment gate SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing payment-gate receipt store fails closed
- **WHEN** `paymentgate.Service.EvaluateDirectPayment` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the unavailable payment-gate receipt store
- **AND** SHALL not silently allow execution or panic

### Requirement: Knowledge runtime dependency guards stay actionable

The knowledge-runtime service SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing knowledgeruntime receipt store fails closed when opening a transaction
- **WHEN** `knowledgeruntime.Service.OpenTransaction` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the `open transaction` action
- **AND** SHALL preserve the actionable `knowledge runtime receipt store is required` cause instead of panicking

#### Scenario: Missing knowledgeruntime receipt store fails closed when selecting an execution path
- **WHEN** `knowledgeruntime.Service.SelectExecutionPath` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the `select execution path` action
- **AND** SHALL preserve the actionable `knowledge runtime receipt store is required` cause instead of panicking

### Requirement: Post-adjudication status dependency guards stay actionable

The post-adjudication status service SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing post-adjudication receipt store fails closed for dead-letter listing
- **WHEN** `postadjudicationstatus.Service.ListCurrentDeadLettersPage` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing post-adjudication receipt store fails closed for transaction status lookup
- **WHEN** `postadjudicationstatus.Service.GetTransactionStatus` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

### Requirement: Settlement progression escalation fails closed on unknown current state

The settlement progression service SHALL return an actionable error instead of panicking when escalation is requested from an unsupported current progression status.

#### Scenario: Escalation from unknown progression status returns an error
- **WHEN** `ApplyReleaseOutcome` maps an `escalate` decision while the current settlement progression status is unknown
- **THEN** the call SHALL return an error identifying the unsupported current settlement progression status
- **AND** SHALL not panic

### Requirement: Post-adjudication replay dependency guards stay actionable

The post-adjudication replay service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing replay receipt store fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing replay dispatcher fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs without a configured dispatcher
- **THEN** the call SHALL return an actionable `dispatcher is required` error instead of panicking

### Requirement: Settlement execution dependency guards stay actionable

The direct settlement execution service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing settlement execution receipt store fails closed
- **WHEN** `settlementexecution.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing settlement execution runtime fails closed
- **WHEN** `settlementexecution.Service.Execute` runs without a configured direct payment runtime
- **THEN** the call SHALL return an actionable `direct payment runtime is required` error instead of panicking

### Requirement: Escrow release dependency guards stay actionable

The escrow release service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing escrow release receipt store fails closed
- **WHEN** `escrowrelease.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing escrow release runtime fails closed
- **WHEN** `escrowrelease.Service.Execute` runs without a configured escrow runtime
- **THEN** the call SHALL return an actionable `escrow runtime is required` error instead of panicking

### Requirement: Escrow refund dependency guards stay actionable

The escrow refund service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing escrow refund receipt store fails closed
- **WHEN** `escrowrefund.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing escrow refund runtime fails closed
- **WHEN** `escrowrefund.Service.Execute` runs without a configured refund runtime
- **THEN** the call SHALL return an actionable `escrow refund runtime is required` error instead of panicking

### Requirement: Dispute hold dependency guards stay actionable

The dispute-hold service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing dispute-hold receipt store fails closed
- **WHEN** `disputehold.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing dispute-hold runtime fails closed
- **WHEN** `disputehold.Service.Execute` runs without a configured dispute-hold runtime
- **THEN** the call SHALL return an actionable `dispute hold runtime is required` error instead of panicking

### Requirement: Escrow adjudication dependency guards stay actionable

The escrow adjudication service SHALL preserve actionable fail-closed behavior when its required receipt store is unavailable.

#### Scenario: Missing escrow adjudication receipt store fails closed
- **WHEN** `escrowadjudication.Service.Adjudicate` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

### Requirement: Partial settlement execution dependency guards stay actionable

The partial settlement execution service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing partial settlement receipt store fails closed
- **WHEN** `partialsettlementexecution.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing partial settlement runtime fails closed
- **WHEN** `partialsettlementexecution.Service.Execute` runs without a configured direct payment runtime
- **THEN** the call SHALL return an actionable `direct payment runtime is required` error instead of panicking

### Requirement: Settlement progression request guards stay actionable

The settlement progression service SHALL preserve actionable validation errors for missing request/store prerequisites.

#### Scenario: Missing settlement progression transaction receipt id fails closed
- **WHEN** `settlementprogression.Service.ApplyReleaseOutcome` runs with an empty `transaction_receipt_id`
- **THEN** the call SHALL return `ErrInvalidApplyReleaseOutcomeRequest`
- **AND** SHALL preserve the actionable `transaction_receipt_id is required` cause

#### Scenario: Missing settlement progression receipt store fails closed
- **WHEN** `settlementprogression.Service.ApplyReleaseOutcome` runs without a configured receipt store
- **THEN** the call SHALL return `ErrInvalidApplyReleaseOutcomeRequest`
- **AND** SHALL preserve the actionable `receipt store is required` cause

### Requirement: Post-adjudication replay request guards stay actionable

The post-adjudication replay service SHALL preserve actionable validation errors for missing replay request identifiers.

#### Scenario: Missing replay transaction receipt id fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs with an empty `transaction_receipt_id`
- **THEN** the call SHALL return an actionable `transaction_receipt_id is required` error instead of panicking

### Requirement: Settlement execution request guards stay actionable

The settlement execution service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing settlement execution transaction receipt id fails closed
- **WHEN** `settlementexecution.Service.Execute` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error instead of panicking

### Requirement: Settlement-cluster execution request-id guards stay consistent

Execution services in the settlement cluster SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Escrow/dispute execution services reject empty transaction receipt ids
- **WHEN** `escrowrelease.Service.Execute`, `escrowrefund.Service.Execute`, `disputehold.Service.Execute`, or `partialsettlementexecution.Service.Execute` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error
- **AND** SHALL not synthesize a denied `missing_receipt` execution result

### Requirement: Mission service request guards stay actionable

The mission service SHALL preserve actionable validation errors for required mission lifecycle inputs.

#### Scenario: Missing accepted-proposal source kind fails closed
- **WHEN** `mission.Service.AcceptProposal` runs without a `source_kind`
- **THEN** the call SHALL return an actionable `accept proposal: source_kind is required` error

#### Scenario: Missing execution reference fails closed
- **WHEN** `mission.Service.AttachExecution` runs without an `execution_ref`
- **THEN** the call SHALL return an actionable `attach execution: execution_ref is required` error

### Requirement: Escrow adjudication request-id guards stay consistent

The escrow adjudication service SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Missing escrow adjudication transaction receipt id fails closed
- **WHEN** `escrowadjudication.Service.Adjudicate` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error
- **AND** SHALL not synthesize a denied `missing_receipt` result

### Requirement: Knowledge runtime request guards stay actionable

The knowledge-runtime service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing knowledge-runtime transaction receipt id fails closed
- **WHEN** `knowledgeruntime.Service.SelectExecutionPath` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction_receipt_id is required` error instead of delegating to downstream store lookup behavior

### Requirement: Post-adjudication status request guards stay actionable

The post-adjudication status service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing post-adjudication transaction receipt id fails closed
- **WHEN** `postadjudicationstatus.Service.GetTransactionStatus` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable transaction-receipt-id-required error instead of reusing the not-found path

### Requirement: Payment gate request-id guards stay consistent

The direct-payment gate SHALL treat an empty transaction receipt id as a validation error instead of a denied business outcome.

#### Scenario: Missing payment-gate transaction receipt id fails closed
- **WHEN** `paymentgate.Service.EvaluateDirectPayment` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable transaction-receipt-id-required error
- **AND** SHALL not synthesize a denied `missing_receipt` result

### Requirement: Escrow execution request-id guards stay actionable

The escrow execution service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing escrow execution transaction receipt id fails closed
- **WHEN** `escrowexecution.Service.ExecuteRecommendation` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error instead of panicking

### Requirement: Knowledge runtime open-transaction request guards stay actionable

The knowledge-runtime service SHALL preserve actionable validation errors for missing canonical open-transaction inputs.

#### Scenario: Missing canonical open inputs fail closed
- **WHEN** `knowledgeruntime.Service.OpenTransaction` runs without the required `transaction_id`, `counterparty`, or `requested_scope`
- **THEN** the call SHALL return `receipts.ErrInvalidSubmissionInput`
- **AND** SHALL preserve the actionable message that those canonical open inputs are required

### Requirement: Meta tool wrapper request guards stay actionable

Transaction-receipt-backed settlement and escrow meta tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Settlement and escrow meta tools reject missing transaction receipt ids
- **WHEN** `execute_settlement`, `execute_partial_settlement`, or `execute_escrow_recommendation` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Broader meta tool wrapper request guards stay actionable

Additional transaction-receipt-backed operator tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Dispute, escrow-release, refund, status, and replay tools reject missing transaction receipt ids
- **WHEN** `hold_escrow_for_dispute`, `release_escrow_settlement`, `refund_escrow_settlement`, `get_post_adjudication_execution_status`, or `retry_post_adjudication_execution` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Remaining meta tool wrapper request guards stay actionable

The remaining transaction-receipt-backed decision and update tools SHALL preserve actionable missing-parameter errors at the wrapper layer.

#### Scenario: Path selection, approval, settlement progression, and escrow adjudication reject missing transaction receipt ids
- **WHEN** `select_knowledge_exchange_path`, `approve_upfront_payment`, `apply_settlement_progression`, or `adjudicate_escrow_dispute` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Decision-tool wrapper outcome guards stay actionable

Decision-oriented transaction-receipt tools SHALL preserve actionable missing-parameter errors for `outcome` at the wrapper layer.

#### Scenario: Settlement progression and escrow adjudication reject missing outcomes
- **WHEN** `apply_settlement_progression` or `adjudicate_escrow_dispute` is invoked without `outcome`
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Canonical open and approval wrapper guards stay actionable

The canonical knowledge-open and upfront-payment-approval tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Open transaction and upfront approval reject missing required inputs
- **WHEN** `open_knowledge_exchange_transaction` or `approve_upfront_payment` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Knowledge-artifact wrapper guards stay actionable

The foundational knowledge-artifact tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Knowledge-artifact tools reject missing required inputs
- **WHEN** `save_knowledge`, `evaluate_exportability`, or `approve_artifact_release` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Learning and skill wrapper guards stay actionable

The learning and skill-management tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Learning and skill tools reject missing required inputs
- **WHEN** `save_learning`, `search_learnings`, or `create_skill` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Skill read/import wrapper guards stay actionable

The skill read/import tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: View and import skill tools reject missing required inputs
- **WHEN** `view_skill` or `import_skill` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: Knowledge read wrapper guards stay actionable

The knowledge history/search tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Knowledge history and search reject missing required inputs
- **WHEN** `get_knowledge_history` or `search_knowledge` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins

### Requirement: P2P payment wrapper request guards stay actionable

The `p2p_pay` tool SHALL preserve actionable missing-parameter errors for its required wrapper inputs before receipt-backed payment evaluation begins.

#### Scenario: Missing p2p-pay required input fails at the wrapper
- **WHEN** `p2p_pay` is invoked without `peer_did`, `transaction_receipt_id`, or `amount`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not defer that validation to the downstream payment gate

### Requirement: Main P2P wrapper guards stay actionable

The main P2P networking tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before session lookup, remote invocation, pricing lookup, or firewall mutation begin.

#### Scenario: Missing P2P required input fails at the wrapper
- **WHEN** `p2p_connect`, `p2p_disconnect`, `p2p_query`, `p2p_firewall_add`, `p2p_firewall_remove`, `p2p_price_query`, `p2p_reputation`, or `p2p_invoke_paid` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Workspace wrapper guards stay actionable

The workspace tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before workspace lookup, membership mutation, or message persistence begin.

#### Scenario: Missing workspace required input fails at the wrapper
- **WHEN** `p2p_workspace_join`, `p2p_workspace_leave`, `p2p_workspace_status`, `p2p_workspace_post`, or `p2p_workspace_read` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Workspace git wrapper guards stay actionable

The workspace git tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before repository lookup, bundle creation, or diff resolution begin.

#### Scenario: Missing workspace git required input fails at the wrapper
- **WHEN** `p2p_git_init`, `p2p_git_push`, `p2p_git_log`, `p2p_git_diff`, or `p2p_git_leaves` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Team workflow wrapper guards stay actionable

Team workflow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Team workflow tools reject missing required inputs
- **WHEN** `team_form`, `team_form_with_budget`, or `team_complete_milestone` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream workflow execution begins

### Requirement: Economy escrow wrapper guards stay actionable

Economy-layer escrow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Economy escrow tools reject missing required inputs
- **WHEN** `economy_escrow_create`, `economy_escrow_milestone`, `economy_escrow_status`, `economy_escrow_release`, or `economy_escrow_dispute` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream escrow execution begins

### Requirement: Core economy wrapper guards stay actionable

Budget, risk, pricing, and negotiation tools SHALL preserve actionable missing-parameter errors for required wrapper inputs before budget lookup, risk assessment, negotiation lookup, or pricing lookup begin.

#### Scenario: Core economy tools reject missing required inputs
- **WHEN** `economy_budget_allocate`, `economy_budget_status`, `economy_budget_close`, `economy_risk_assess`, `economy_negotiate`, `economy_negotiate_status`, or `economy_price_quote` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream economy execution begins

### Requirement: Ontology governance/action wrapper guards stay actionable

Ontology governance and dynamic action tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Ontology governance and action tools reject missing required inputs
- **WHEN** `ontology_promote_type`, `ontology_promote_predicate`, `ontology_type_usage`, or any `ontology_action_*` tool is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream ontology execution begins

### Requirement: On-chain escrow wrapper guards stay actionable

On-chain escrow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: On-chain escrow tools reject missing required inputs
- **WHEN** `escrow_create`, `escrow_fund`, `escrow_activate`, `escrow_submit_work`, `escrow_release`, `escrow_refund`, `escrow_dispute`, `escrow_resolve`, or `escrow_status` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream escrow execution begins
- **AND** this SHALL cover `escrow_create` inputs such as `buyerDid`, `sellerDid`, `amount`, and `milestones`, plus `escrow_resolve` inputs such as `escrowId`, `favor`, and `sellerPercent`

### Requirement: Payment send wrapper guards stay actionable

The `payment_send` tool SHALL preserve actionable missing-parameter errors for its required wrapper inputs before receipt-backed payment evaluation begins.

#### Scenario: Missing payment-send transaction receipt id fails at the wrapper
- **WHEN** `payment_send` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not defer that validation to the downstream payment gate

### Requirement: X402 fetch wrapper guards stay actionable

The `payment_x402_fetch` tool SHALL preserve actionable missing-parameter errors for its required wrapper inputs before network request construction begins.

#### Scenario: Missing X402 fetch URL fails at the wrapper
- **WHEN** `payment_x402_fetch` is invoked without `url`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Browser action conditional guards stay actionable

The `browser_action` tool SHALL preserve actionable validation errors for action-specific required inputs before DOM interaction or script evaluation begins.

#### Scenario: Missing browser action selector or text fails before interaction
- **WHEN** `browser_action` is invoked without the `selector` or `text` required by the chosen action
- **THEN** the tool SHALL return an actionable validation error
- **AND** SHALL not defer that failure to downstream browser interaction

### Requirement: Browser search and navigation guards stay actionable

The browser search and navigation entrypoints SHALL preserve actionable missing-parameter errors for required top-level inputs before session creation or navigation begins.

#### Scenario: Missing browser query or URL fails at the wrapper
- **WHEN** `browser_search` or `browser_navigate` is invoked without its required `query` or `url`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Web retrieval wrapper guards stay actionable

The lightweight web retrieval tools SHALL preserve actionable missing-parameter errors for required top-level inputs before any HTTP request begins.

#### Scenario: Missing web search or fetch input fails at the wrapper
- **WHEN** `web_search` or `web_fetch` is invoked without its required `query` or `url`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Background wrapper guards stay actionable

The background-task tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before queue submission or task lookup begins.

#### Scenario: Missing background prompt or task id fails at the wrapper
- **WHEN** `bg_submit`, `bg_status`, `bg_result`, or `bg_cancel` is invoked without its required `prompt` or `task_id`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Workflow wrapper guards stay actionable

The workflow tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before workflow lookup, cancellation, or file writes begin.

#### Scenario: Missing workflow run id or save input fails at the wrapper
- **WHEN** `workflow_status`, `workflow_cancel`, or `workflow_save` is invoked without its required `run_id`, `name`, or `yaml_content`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Cron wrapper guards stay actionable

The cron tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before scheduler lookup or mutation begins.

#### Scenario: Missing cron add or control input fails at the wrapper
- **WHEN** `cron_add`, `cron_pause`, `cron_resume`, or `cron_remove` is invoked without its required `name`, `schedule_type`, `schedule`, `prompt`, or `id`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Graph and agent-memory wrapper guards stay actionable

The graph and agent-memory tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before traversal, lookup, or mutation begins.

#### Scenario: Missing graph or agent-memory input fails at the wrapper
- **WHEN** `graph_traverse`, `graph_query`, `memory_agent_save`, `memory_agent_recall`, or `memory_agent_forget` is invoked without its required input
- **THEN** the tool SHALL return an actionable validation error

### Requirement: Librarian inquiry wrapper guards stay actionable

The librarian inquiry tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before inquiry lookup or mutation begins.

#### Scenario: Missing inquiry id fails at the wrapper
- **WHEN** `librarian_dismiss_inquiry` is invoked without `inquiry_id`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Output retrieval wrapper guards stay actionable

The output retrieval tool SHALL preserve actionable validation errors for required wrapper inputs before stored-output lookup begins.

#### Scenario: Missing output ref or grep pattern fails at the wrapper
- **WHEN** `tool_output_get` is invoked without `ref`, or with mode `grep` but without `pattern`
- **THEN** the tool SHALL return an actionable validation error

### Requirement: Run-ledger wrapper guards stay actionable

The `run_*` control-plane tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before journal writes, snapshot reads, or policy application begins.

#### Scenario: Missing run-ledger input fails at the wrapper
- **WHEN** `run_create`, `run_read`, `run_active`, `run_note`, `run_propose_step_result`, `run_apply_policy`, `run_approve_step`, or `run_resume` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Agent control-plane wrapper guards stay actionable

The `agentrt` control-plane and task-management tools SHALL preserve actionable missing-parameter errors for required wrapper inputs before run creation, run-store lookup, cancellation, or task-store mutation begins.

#### Scenario: Missing agent or task required input fails at the wrapper
- **WHEN** `agent_spawn`, `agent_wait`, `agent_stop`, `task_create`, `task_get`, or `task_update` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Exec wrapper guards stay actionable

The exec tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before policy evaluation or supervisor interaction begins.

#### Scenario: Missing exec command or id fails at the wrapper
- **WHEN** `exec`, `exec_bg`, `exec_status`, or `exec_stop` is invoked without its required `command` or `id`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Vault crypto and secrets input guards stay actionable

The vault crypto and secrets tools SHALL preserve actionable validation errors for required inputs before key lookup, secret-store access, or cryptographic execution begins.

#### Scenario: Missing crypto or secrets required input fails before execution
- **WHEN** `crypto_encrypt`, `crypto_decrypt`, `crypto_sign`, `crypto_hash`, `secrets_store`, `secrets_get`, or `secrets_delete` is invoked without its required input
- **THEN** the tool SHALL return an actionable validation error

### Requirement: Contract tool wrapper guards stay actionable

The contract tool cluster SHALL preserve actionable missing-parameter errors for required wrapper inputs before ABI parsing, RPC reads, cache mutation, or transaction submission begins.

#### Scenario: Missing contract required input fails at the wrapper
- **WHEN** `contract_read`, `contract_call`, or `contract_abi_load` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Sentinel acknowledge wrapper guard stays actionable

The Security Sentinel acknowledge tool SHALL preserve an actionable missing-parameter error for `alertId` before alert-store mutation begins.

#### Scenario: Missing sentinel alert id fails at the wrapper
- **WHEN** `sentinel_acknowledge` is invoked without `alertId`
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Smart-account wrapper guards stay actionable

The smart-account session, policy, module, and paymaster tools SHALL preserve actionable missing-parameter errors for required wrapper inputs before session creation, policy validation, module mutation, or approval submission begins.

#### Scenario: Missing smart-account required input fails at the wrapper
- **WHEN** `session_key_create`, `session_key_revoke`, `session_execute`, `policy_check`, `module_install`, `module_uninstall`, or `paymaster_approve` is invoked without one of its required wrapper inputs
- **THEN** the tool SHALL return an actionable missing-parameter error

### Requirement: Production Go code avoids placeholder contexts
Production Go code SHALL not contain `context.TODO()` placeholders in non-test source files.

#### Scenario: Production Go files contain no context.TODO calls
- **WHEN** repository quality guard tests scan non-test Go files under `cmd/` and `internal/`
- **THEN** the scan SHALL find zero `context.TODO()` occurrences

### Requirement: Internal CLI packages avoid direct process exits

Non-test Go files under `internal/cli/` SHALL NOT call or assign `os.Exit` directly. CLI packages that need command-specific non-zero status codes SHALL return structured errors to the binary entrypoint, and only `cmd/*/main.go` SHALL terminate the process.

#### Scenario: Internal CLI package exit hygiene
- **WHEN** repository quality tests scan non-test Go files under `internal/cli/`
- **THEN** the scan SHALL find zero direct `os.Exit` references

#### Scenario: Extension CLI returns exit-code errors
- **WHEN** an extension command needs to signal exit code 1, 2, or 3
- **THEN** it SHALL return a structured CLI error carrying that code
- **AND** it SHALL NOT terminate the process itself
