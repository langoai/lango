## Purpose

Capability spec for meta-tools. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Knowledge Management Tools
The system SHALL provide agent-facing tools for managing the knowledge base.

#### Scenario: save_knowledge tool
- **WHEN** the agent invokes `save_knowledge` with key, category, content, and optional tags/source
- **THEN** the system SHALL validate the category using `entknowledge.CategoryValidator()` before persisting
- **AND** persist the knowledge entry via the Store
- **AND** create an audit log entry with action "knowledge_save"
- **AND** return a success status with the key

#### Scenario: save_knowledge with invalid category
- **WHEN** the agent invokes `save_knowledge` with an unrecognized category
- **THEN** the system SHALL return an error indicating the invalid category without saving

#### Scenario: save_knowledge tool schema includes all categories
- **WHEN** the tool parameters are inspected
- **THEN** the `category` enum SHALL include all six valid values: rule, definition, preference, fact, pattern, correction

#### Scenario: search_knowledge tool
- **WHEN** the agent invokes `search_knowledge` with a query and optional category
- **THEN** the system SHALL search knowledge entries via the Store
- **AND** return matching results with count

### Requirement: Exportability evaluation tool
The meta tools surface SHALL provide an `evaluate_exportability` tool that evaluates an artifact label and a list of knowledge source keys, then returns the exportability decision receipt.

#### Scenario: Tool returns receipt payload
- **WHEN** `evaluate_exportability` is invoked with `artifact_label`, `source_keys`, and `stage`
- **THEN** it SHALL load the latest knowledge entries for the provided keys
- **AND** it SHALL evaluate them through the exportability policy engine
- **AND** it SHALL return the decision stage, decision state, policy code, explanation, and lineage summary

#### Scenario: Tool emits audit-backed exportability receipt
- **WHEN** `evaluate_exportability` completes
- **THEN** it SHALL append an `exportability_decision` audit entry recording the receipt payload

### Requirement: Artifact release approval tool
The meta tools surface SHALL provide an `approve_artifact_release` tool that evaluates a release request and returns a structured approval outcome.

#### Scenario: Tool returns structured approval outcome
- **WHEN** `approve_artifact_release` is invoked with artifact label, requested artifact label, and exportability state
- **THEN** it SHALL evaluate the request through the approval-flow domain model
- **AND** it SHALL return the decision, reason, issue classification, fulfillment assessment, and settlement hint

#### Scenario: Tool emits audit-backed approval receipt
- **WHEN** `approve_artifact_release` completes
- **THEN** it SHALL append an `artifact_release_approval` audit entry recording the approval outcome

### Requirement: Dispute-ready receipt creation tool
The meta tools surface SHALL provide a `create_dispute_ready_receipt` tool that creates a lite submission receipt and links it to a transaction receipt.

#### Scenario: Tool creates submission and transaction linkage
- **WHEN** `create_dispute_ready_receipt` is invoked with transaction ID, artifact label, payload hash, and source lineage digest
- **THEN** it SHALL create a submission receipt
- **AND** it SHALL create or reuse the corresponding transaction receipt
- **AND** it SHALL return the created submission receipt ID, transaction receipt ID, and current submission pointer

### Requirement: Upfront payment approval tool
The meta tools surface SHALL provide an `approve_upfront_payment` tool that evaluates a prepayment request and records the result.

#### Scenario: Tool returns approval outcome
- **WHEN** `approve_upfront_payment` is invoked with transaction receipt ID, amount, trust input, and budget or policy context
- **THEN** it SHALL evaluate the request through the upfront payment approval domain model
- **AND** it SHALL return the decision, reason, suggested payment mode, amount class, and risk class

#### Scenario: Tool updates transaction receipt
- **WHEN** `approve_upfront_payment` completes
- **THEN** it SHALL update the linked transaction receipt with canonical payment approval state and append the corresponding event

### Requirement: Learning Management Tools
The system SHALL provide agent-facing tools for managing learned patterns.

#### Scenario: save_learning tool
- **WHEN** the agent invokes `save_learning` with trigger, fix, and optional error_pattern/diagnosis/category
- **THEN** the system SHALL persist the learning entry via the Store
- **AND** create an audit log entry with action "learning_save"
- **AND** return a success status

#### Scenario: search_learnings tool
- **WHEN** the agent invokes `search_learnings` with a query and optional category
- **THEN** the system SHALL search learning entries via the Store
- **AND** return matching results with count

### Requirement: Skill Management Tools
The system SHALL provide agent-facing tools for creating and listing skills.

#### Scenario: create_skill tool
- **WHEN** the agent invokes `create_skill` with name, description, type, and definition (JSON string)
- **THEN** the system SHALL parse the definition JSON
- **AND** create the skill via the Registry
- **AND** if auto-approve is enabled, SHALL activate the skill immediately
- **AND** create an audit log entry with action "skill_create"
- **AND** return the skill status ("draft" or "active")

#### Scenario: list_skills tool
- **WHEN** the agent invokes `list_skills`
- **THEN** the system SHALL return all active skills with their metadata

### Requirement: Tool Learning Wrapper
The system SHALL wrap existing tool handlers to feed execution results into the learning engine.

#### Scenario: Wrap tool with learning
- **WHEN** `wrapWithLearning` is called on a tool
- **THEN** the system SHALL return a new tool with the same name, description, and parameters
- **AND** the wrapped handler SHALL call the original handler first
- **AND** then call `engine.OnToolResult` with the tool name, params, result, and error
- **AND** return the original result and error unchanged

### Requirement: Skill import access control
The `import_skill` tool handler SHALL check `SkillConfig.AllowImport` before processing any import request. When `AllowImport` is false, the handler SHALL return an error indicating skill import is disabled.

#### Scenario: Import blocked when AllowImport is false
- **WHEN** `import_skill` is invoked and `SkillConfig.AllowImport` is `false`
- **THEN** the handler SHALL return error "skill import disabled (skill.allowImport=false)"
- **AND** no import processing SHALL occur

#### Scenario: Import proceeds when AllowImport is true
- **WHEN** `import_skill` is invoked and `SkillConfig.AllowImport` is `true`
- **THEN** the handler SHALL proceed with normal import logic

### Requirement: Save knowledge meta tool
The `save_knowledge` tool SHALL accept exportability-related source tagging metadata in addition to key, category, content, tags, and source.

#### Scenario: Save knowledge with source tagging
- **WHEN** `save_knowledge` is called with `source_class` and `asset_label`
- **THEN** the stored knowledge entry SHALL persist those fields for later exportability evaluation

#### Scenario: Save knowledge default source class
- **WHEN** `save_knowledge` is called without `source_class`
- **THEN** the tool SHALL default the stored source class to `private-confidential`

### Requirement: Knowledge exchange runtime control plane reuses receipt-backed meta tools
The meta tools surface SHALL treat the first knowledge exchange runtime design slice as a composition of the existing receipt-backed tools, with `transaction receipt` as canonical control-plane state and `submission receipt` as canonical deliverable state.

#### Scenario: Runtime slice reuses existing tool contracts
- **WHEN** the knowledge exchange runtime slice is described through meta-tools behavior
- **THEN** it SHALL rely on the existing exportability, approval, submission-creation, upfront-payment, and escrow recommendation tools rather than introducing a duplicate receipt model

### Requirement: Settlement progression meta tool
The meta tools surface SHALL provide an `apply_settlement_progression` tool that maps artifact release outcomes into transaction-level settlement progression state.

#### Scenario: Settlement progression tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `apply_settlement_progression` SHALL be available

#### Scenario: Settlement progression tool applies release outcomes
- **WHEN** `apply_settlement_progression` is invoked with `transaction_receipt_id`, `outcome`, and optional `reason` or `partial_hint`
- **THEN** it SHALL evaluate the request through the settlement progression service
- **AND** it SHALL return canonical transaction-level settlement progression state

### Requirement: Actual settlement execution meta tool
The meta tools surface SHALL provide an `execute_settlement` tool that executes direct final settlement from canonical settlement progression state.

#### Scenario: Actual settlement execution tool available
- **WHEN** the meta tools are built with a receipts store and settlement runtime
- **THEN** `execute_settlement` SHALL be available

#### Scenario: Actual settlement execution tool executes direct settlement
- **WHEN** `execute_settlement` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the settlement execution service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state

### Requirement: Partial settlement execution meta tool
The meta tools surface SHALL provide an `execute_partial_settlement` tool that executes a direct partial settlement from canonical settlement progression state.

#### Scenario: Partial settlement execution tool available
- **WHEN** the meta tools are built with a receipts store and partial-settlement runtime
- **THEN** `execute_partial_settlement` SHALL be available

#### Scenario: Partial settlement execution tool executes direct partial settlement
- **WHEN** `execute_partial_settlement` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the partial settlement execution service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state, executed amount, remaining amount, and runtime reference

### Requirement: Escrow release meta tool
The meta tools surface SHALL provide a `release_escrow_settlement` tool that executes funded escrow release from canonical settlement progression state.

#### Scenario: Escrow release tool available
- **WHEN** the meta tools are built with a receipts store and escrow release runtime
- **THEN** `release_escrow_settlement` SHALL be available

#### Scenario: Escrow release tool executes funded release
- **WHEN** `release_escrow_settlement` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the escrow release service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state, resolved amount, and runtime reference

### Requirement: Escrow refund meta tool
The meta tools surface SHALL provide a `refund_escrow_settlement` tool that executes funded escrow refund from canonical review-needed settlement state.

#### Scenario: Escrow refund tool available
- **WHEN** the meta tools are built with a receipts store and escrow refund runtime
- **THEN** `refund_escrow_settlement` SHALL be available

#### Scenario: Escrow refund tool executes funded refund
- **WHEN** `refund_escrow_settlement` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the escrow refund service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state, resolved amount, and runtime reference

### Requirement: Dispute hold meta tool
The meta tools surface SHALL provide a `hold_escrow_for_dispute` tool that records dispute hold evidence for funded escrow from canonical dispute-ready settlement state.

#### Scenario: Dispute hold tool available
- **WHEN** the meta tools are built with a receipts store and dispute hold runtime
- **THEN** `hold_escrow_for_dispute` SHALL be available

#### Scenario: Dispute hold tool records funded dispute hold
- **WHEN** `hold_escrow_for_dispute` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the dispute hold service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state, escrow reference, and runtime reference

### Requirement: Release vs refund adjudication meta tool
The meta tools surface SHALL provide an `adjudicate_escrow_dispute` tool that records the first canonical release-vs-refund branch after dispute hold.

#### Scenario: Adjudication tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `adjudicate_escrow_dispute` SHALL be available

#### Scenario: Adjudication tool records release or refund branch
- **WHEN** `adjudicate_escrow_dispute` is invoked with `transaction_receipt_id` and `outcome`
- **THEN** it SHALL evaluate the request through the escrow adjudication service
- **AND** it SHALL atomically record the adjudication field and the corresponding settlement progression transition
- **AND** it SHALL return canonical transaction-level adjudication result including settlement progression state, escrow reference, and outcome

#### Scenario: Adjudication tool may inline nested execution
- **WHEN** `adjudicate_escrow_dispute` is invoked with `auto_execute=true`
- **THEN** it SHALL preserve adjudication as the canonical write layer
- **AND** it SHALL, after adjudication succeeds, invoke the matching release or refund executor inline
- **AND** it SHALL return both the adjudication result and the nested execution result when available

#### Scenario: Adjudication tool may enqueue background execution
- **WHEN** `adjudicate_escrow_dispute` is invoked with `background_execute=true`
- **THEN** it SHALL preserve adjudication as the canonical write layer
- **AND** it SHALL enqueue the matching release or refund follow-up onto the background task substrate
- **AND** it SHALL return both the adjudication result and a background dispatch receipt

#### Scenario: Background post-adjudication execution uses bounded retry
- **WHEN** background post-adjudication execution fails
- **THEN** the post-adjudication path SHALL retry up to three times with exponential backoff
- **AND** exhausted retries SHALL produce terminal dead-letter evidence without changing canonical adjudication

### Requirement: Operator replay tool
The meta tools surface SHALL provide a `retry_post_adjudication_execution` tool that replays dead-lettered post-adjudication execution through the existing background dispatch path.

#### Scenario: Replay tool available
- **WHEN** the meta tools are built with receipts and background dispatch support
- **THEN** `retry_post_adjudication_execution` SHALL be available

#### Scenario: Replay tool returns adjudication snapshot and dispatch receipt
- **WHEN** `retry_post_adjudication_execution` succeeds
- **THEN** it SHALL return the canonical adjudication snapshot and the new background dispatch receipt

#### Scenario: Replay tool enforces actor-based policy
- **WHEN** `retry_post_adjudication_execution` is invoked
- **THEN** it SHALL fail closed when actor resolution fails
- **AND** it SHALL fail closed when the actor is not allowed for the current replay outcome

### Requirement: Dead-letter browsing and status observation tools
The meta tools surface SHALL provide read-only visibility into dead-lettered post-adjudication execution.

#### Scenario: Dead-letter backlog tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `list_dead_lettered_post_adjudication_executions` SHALL be available

#### Scenario: Post-adjudication status tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `get_post_adjudication_execution_status` SHALL be available

#### Scenario: Dead-letter backlog tool supports filtering and pagination
- **WHEN** `list_dead_lettered_post_adjudication_executions` is invoked
- **THEN** it SHALL accept `adjudication`, `retry_attempt_min`, `retry_attempt_max`, `query`, `manual_replay_actor`, `dead_lettered_after`, `dead_lettered_before`, `dead_letter_reason_query`, `latest_dispatch_reference`, `latest_status_subtype`, `manual_retry_count_min`, `manual_retry_count_max`, `total_retry_count_min`, `total_retry_count_max`, `transaction_global_total_retry_count_min`, `transaction_global_total_retry_count_max`, `transaction_global_any_match_family`, `latest_status_subtype_family`, `any_match_family`, `dominant_family`, `sort_by`, `offset`, and `limit`
- **AND** it SHALL return `entries`, `count`, `total`, `offset`, and `limit`
- **AND** each entry SHALL expose `latest_dead_lettered_at`, `latest_manual_replay_actor`, `latest_manual_replay_at`, `latest_status_subtype`, `manual_retry_count`, `total_retry_count`, `transaction_global_total_retry_count`, `transaction_global_any_match_families`, `latest_status_subtype_family`, `any_match_families`, and `dominant_family`

#### Scenario: Post-adjudication status tool returns navigation hints
- **WHEN** `get_post_adjudication_execution_status` succeeds
- **THEN** it SHALL return the current canonical snapshot
- **AND** it SHALL return the latest retry / dead-letter summary
- **AND** it SHALL return `is_dead_lettered`, `can_retry`, and `adjudication`

### Requirement: Escrow release and refund meta tools enforce canonical adjudication
The meta tools surface SHALL enforce canonical adjudication on the existing escrow release and refund execution tools.

#### Scenario: Release tool requires matching adjudication
- **WHEN** `release_escrow_settlement` is invoked
- **THEN** it SHALL require `escrow_adjudication = release`
- **AND** it SHALL deny execution when adjudication is missing or mismatched

#### Scenario: Refund tool requires matching adjudication
- **WHEN** `refund_escrow_settlement` is invoked
- **THEN** it SHALL require `escrow_adjudication = refund`
- **AND** it SHALL deny execution when adjudication is missing or mismatched
