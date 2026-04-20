# Trust, Security & Policy Audit

## Purpose

This document is the second detailed audit ledger under the Lango master document.

It exists to review the capability area that determines whether `knowledge exchange v1` can operate safely and coherently:

- identity and auth entry,
- privacy and output boundaries,
- approval and execution control,
- auditability and cryptographic accountability.

## Relationship to the Master Document

This audit sits underneath `docs/architecture/master-document.md` and must use that document's constitution, capability taxonomy, audit vocabulary, and track-routing rules.

It does not redefine what Lango is, replace the product constitution, or create new top-level capability areas or execution tracks. Its role is to apply the master document's framework to the `Trust, Security & Policy` capability area in detailed ledger form, using `knowledge exchange v1` as the judgment baseline.

## Document Ownership

- Primary capability area: `Trust, Security & Policy`
- Primary execution track: `Stabilization Track`
- Secondary capability areas:
  - `External Collaboration & Economic Exchange`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `P2P Knowledge Exchange Track`

## Audit Order

1. Identity, Auth & Trust Entry
2. Privacy, Exportability & Output Policy
3. Approval, Execution Policy & Sandboxing
4. Auditability, Provenance & Cryptographic Accountability

## Audit Method

This ledger adopts the master document's minimum audit schema for detailed row-level work.

Each row must include:

- feature name,
- capability area,
- product-path linkage,
- current surface area,
- core value,
- current problem,
- judgment,
- execution track,
- secondary capability areas,
- secondary tracks.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

The judgment baseline for this audit is narrow by design:

- Does this capability make `knowledge exchange v1` safer and clearer?
- Does it create a usable external-market boundary for early deliverable exchange?
- Does it create operator-visible confidence without over-claiming what the runtime can actually enforce?

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| Identity, Auth & Trust Entry | Phase 1-2 | `docs/security/authentication.md`, `docs/gateway/http-api.md`, `internal/gateway/auth.go`, `internal/p2p/handshake/*`, `internal/config/types_security.go` | Detailed audit complete (`stabilize`) |
| Privacy, Exportability & Output Policy | Phase 1 | `docs/security/index.md`, `docs/security/exportability.md`, `docs/security/pii-redaction.md`, `internal/cli/security/status.go`, `internal/config/types_security.go`, `internal/config/types.go`, `internal/gatekeeper/*` | Detailed audit complete (`stabilize`) |
| Approval, Execution Policy & Sandboxing | Phase 1-2 | `docs/security/tool-approval.md`, `docs/security/approval-flow.md`, `docs/security/index.md`, `docs/cli/sandbox.md`, `internal/toolchain/mw_approval.go`, `internal/approvalflow/*`, `internal/app/tools_meta_approvalflow.go`, `internal/tools/exec/*`, `internal/sandbox/*`, `internal/cli/settings/forms_security.go`, `internal/cli/settings/forms_sandbox.go` | Detailed audit complete (`stabilize`) |
| Auditability, Provenance & Cryptographic Accountability | Phase 1-2 | `docs/features/provenance.md`, `docs/security/dispute-ready-receipts.md`, `docs/security/encryption.md`, `internal/observability/audit/*`, `internal/provenance/*`, `internal/security/*`, `internal/app/wiring_provenance.go` | Detailed audit complete (`stabilize`); lite dispute-ready receipt model landed for submission/transaction records and event trails |

## Baseline Decisions Already Locked

- `knowledge exchange v1` is deliverable-oriented by default, not unrestricted remote execution.
- Private conversations, confidential material, and raw sensitive inputs are not tradeable by default.
- Tradeable knowledge and deliverables must stay inside an allowlist plus explicit exportability policy.
- Limited execution opens only under higher trust, explicit approval, and stronger policy boundaries.
- Dispute handling for early exchange should be grounded in signed logs, provenance, acceptance criteria, escrow state, and immutable receipts.

## Detailed Audit: Identity, Auth & Trust Entry

### Audit Record

- Feature name: `Identity, Auth & Trust Entry`
- Capability area: `Trust, Security & Policy`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/security/authentication.md`, `docs/gateway/http-api.md`, `internal/gateway/auth.go`, `internal/p2p/handshake/security_events.go`, `internal/config/types_security.go`
- Core value: `Define who is allowed to cross the gateway boundary, who is allowed to maintain a P2P session, and how trust entry is revoked when conditions deteriorate.`
- Current problem: `Auth entry and P2P trust entry are both real, but they still read like parallel subsystems instead of one canonical external-market entry model.`
- Judgment: `stabilize`
- Execution track: `Stabilization Track`
- Secondary capability areas:
  - `External Collaboration & Economic Exchange`
- Secondary tracks:
  - `P2P Knowledge Exchange Track`

### Findings

1. `Major` Gateway authentication is a real capability with meaningful hardening, but it is still documented as a gateway concern rather than an external-market entry model.
   - OIDC login is rate-limited, uses per-provider state cookies, clears the state cookie after validation, and writes a `lango_session` cookie after creating a session keyed only by provider and subject.
   - The callback returns a structured JSON response without echoing user email addresses.
   - References: `internal/gateway/auth.go:84-121`, `internal/gateway/auth.go:132-210`, `docs/security/authentication.md:7-46`, `docs/security/authentication.md:115-120`

2. `Major` Protected-route documentation is still fragmented across gateway auth and P2P docs.
   - `docs/security/authentication.md` presents `/ws` and `/status` as the protected route story.
   - `docs/gateway/http-api.md` separately explains that `/api/p2p/*` is also gated when OIDC is configured.
   - The runtime behavior is acceptable, but the operator-facing entry story is split.
   - References: `docs/security/authentication.md:65-84`, `docs/gateway/http-api.md:59-62`

3. `Major` P2P trust entry and revocation are real, but they are still expressed as security-event mechanics rather than one operator-facing trust boundary.
   - The handshake security handler auto-invalidates sessions after repeated tool failures or reputation drops below threshold.
   - That is a meaningful trust-entry control, but it lives beside gateway auth rather than inside one unified model for early external exchange.
   - References: `internal/p2p/handshake/security_events.go:9-17`, `internal/p2p/handshake/security_events.go:20-37`, `internal/p2p/handshake/security_events.go:40-73`

### Assessment

- `Identity, Auth & Trust Entry` is a real capability and should be kept.
- The right action is `stabilize`, not `merge` or `defer`:
  - publish one canonical operator-facing entry model for gateway auth plus P2P trust entry,
  - make the protected-route story read as one surface,
  - connect session invalidation and trust revocation to the external-market narrative instead of leaving them as internal mechanics.

## Detailed Audit: Privacy, Exportability & Output Policy

### Audit Record

- Feature name: `Privacy, Exportability & Output Policy`
- Capability area: `Trust, Security & Policy`
- Product-path linkage: `Phase 1: Knowledge Exchange`
- Current surface area: `docs/security/index.md`, `docs/security/pii-redaction.md`, `internal/config/types_security.go`, `internal/config/types.go`, `internal/gatekeeper/sanitizer.go`
- Core value: `Prevent sensitive inputs and unsafe outputs from leaking into external exchange, while defining what is and is not tradeable in early knowledge exchange.`
- Current problem: `Privacy tooling exists, and exportability now has a first source-primary runtime/operator surface, but the broader policy model is still intentionally narrow, so hygiene and tradeability are no longer conflated.`
- Judgment: `stabilize`
- Execution track: `Stabilization Track`
- Secondary capability areas:
  - `External Collaboration & Economic Exchange`
- Secondary tracks:
  - `P2P Knowledge Exchange Track`

### Findings

1. `Major` Input and output hygiene are real capabilities.
   - The security interceptor exposes PII redaction, built-in/custom regex controls, and optional Presidio analysis.
   - The gatekeeper sanitizer independently strips thought tags, internal markers, and oversized raw JSON, and supports custom removal patterns.
   - References: `docs/security/index.md:9-23`, `docs/security/pii-redaction.md:7-27`, `docs/security/pii-redaction.md:71-181`, `internal/config/types_security.go:104-125`, `internal/config/types.go:529-548`, `internal/gatekeeper/sanitizer.go:11-18`, `internal/gatekeeper/sanitizer.go:21-37`, `internal/gatekeeper/sanitizer.go:45-84`

2. `Major` Exportability policy is still a product requirement, but the first runtime/operator slice is now landed.
   - The master product path and P2P knowledge-exchange track require explicit exportability policy.
   - The actual security config now exposes exportability alongside approval, PII, and Presidio settings, and `lango security status` surfaces the current state.
   - In practice, the runtime can now answer the narrow first-slice question from source lineage, but it still cannot express a full policy-rule DSL, human override UI, or unified dispute-ready receipt model.
   - References: `docs/architecture/master-document.md:202-204`, `docs/architecture/p2p-knowledge-exchange-track.md:46-51`, `internal/config/types_security.go:104-125`, `internal/config/types.go:529-548`

3. `Major` The current output policy is hygiene-oriented, not market-oriented.
   - Gatekeeper strips internal tags, marker-prefixed lines, and large JSON payloads.
   - That is useful for output cleanliness, but it is not a policy system for derived knowledge, confidential source material, or explicit tradeability decisions.
   - References: `internal/gatekeeper/sanitizer.go:25-28`, `internal/gatekeeper/sanitizer.go:51-84`

### Assessment

- `Privacy, Exportability & Output Policy` should be kept and strengthened.
- The correct action is `stabilize`:
  - preserve the existing PII and gatekeeper layers,
  - stop treating sanitization as if it already solved exportability,
  - add an explicit policy design for tradeable vs non-tradeable artifacts in the next planning cycle.

## Detailed Audit: Approval, Execution Policy & Sandboxing

### Audit Record

- Feature name: `Approval, Execution Policy & Sandboxing`
- Capability area: `Trust, Security & Policy`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/security/tool-approval.md`, `internal/toolchain/mw_approval.go`, `internal/cli/settings/forms_security.go`, `internal/tools/exec/exec.go`, `internal/tools/exec/policy.go`, `internal/cli/settings/forms_sandbox.go`, `internal/sandbox/*`
- Core value: `Control what execution is allowed, when explicit approval is required, and whether limited execution runs inside meaningful isolation boundaries.`
- Current problem: `Approval and sandboxing are both real, but the current operator model still mixes strong controls with deliberate bypass paths, so limited execution is not yet one clean Phase 2 story.`
- Judgment: `stabilize`
- Execution track: `Stabilization Track`
- Secondary capability areas:
  - `External Collaboration & Economic Exchange`
- Secondary tracks:
  - `P2P Knowledge Exchange Track`

### Findings

1. `Major` Approval is already a substantial execution-control subsystem.
   - The middleware uses fail-closed semantics unless a request is explicitly approved.
   - It supports turn-local approval reuse, persistent session grants, bounded timeout retries, approval history, and auto-approval for small payments through the spending limiter.
   - References: `docs/security/tool-approval.md:7-18`, `docs/security/tool-approval.md:83-107`, `internal/toolchain/mw_approval.go:17-23`, `internal/toolchain/mw_approval.go:43-76`, `internal/toolchain/mw_approval.go:91-118`, `internal/toolchain/mw_approval.go:123-156`, `internal/toolchain/mw_approval.go:157-259`

2. `Major` Approval still includes explicit bypass modes that must be treated as policy exceptions, not invisible defaults.
   - `approvalPolicy: none` disables approval entirely.
   - `headlessAutoApprove` is documented as bypassing the workflow.
   - `exemptTools` override both the global policy and configured sensitive-tool lists.
   - These are legitimate controls, but they mean `limited execution` is only conditionally policy-bound today.
   - References: `docs/security/tool-approval.md:13-23`, `docs/security/tool-approval.md:59-81`, `docs/security/tool-approval.md:132-150`, `internal/config/types_security.go:104-117`, `internal/cli/settings/forms_security.go:35-40`, `internal/cli/settings/forms_security.go:72-77`

3. `Major` OS sandboxing is real, but the current execution story still allows explicit unsandboxed paths.
   - The exec tool can reject execution when `failClosed` is enabled, but otherwise proceeds unsandboxed with a warning when no isolator is available.
   - `ExcludedCommands` bypass the sandbox entirely and are only recorded in audit.
   - The settings surface is honest about this, including the fact that excluded commands run unsandboxed.
   - References: `internal/tools/exec/exec.go:105-141`, `internal/tools/exec/exec.go:163-170`, `internal/cli/settings/forms_sandbox.go:24-40`, `internal/cli/settings/forms_sandbox.go:79-84`

4. `Major` The sandbox surface still spans multiple operator planes.
   - There is OS-level sandboxing for local tool execution and a separate P2P tool-isolation/container path elsewhere in the system.
   - For `knowledge exchange v1` that is acceptable, but for `limited execution Phase 2` it will need one clearer operator story.
   - References: `internal/cli/settings/forms_sandbox.go:17-19`, `internal/cli/settings/forms_sandbox.go:39-40`, `internal/tools/exec/exec.go:105-141`, `internal/cli/p2p/sandbox.go:16-18`

### Assessment

- Post-implementation note: the first-slice artifact release approval path is now landed with structured `approve`, `reject`, `request-revision`, and `escalate` outcomes plus audit-backed receipts.
- The remaining gaps are still explicit: no human approval UI, no upfront payment approval runtime, no dispute orchestration, and no partial settlement execution.

- `Approval, Execution Policy & Sandboxing` is a real capability and should be kept.
- The correct action is `stabilize`:
  - make the bypass cases explicit policy exceptions,
  - define what `limited execution` requires in Phase 2,
  - connect approval and sandbox controls into one operator-facing execution model.

## Detailed Audit: Auditability, Provenance & Cryptographic Accountability

### Audit Record

- Feature name: `Auditability, Provenance & Cryptographic Accountability`
- Capability area: `Trust, Security & Policy`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/provenance.md`, `internal/observability/audit/recorder.go`, `internal/provenance/bundle.go`, `internal/app/wiring_provenance.go`, `internal/security/kms_factory.go`, `internal/config/types_security.go`
- Core value: `Create durable evidence about what happened, who signed it, what was redacted, and what can later be used in approval, settlement, or dispute handling.`
- Current problem: `The building blocks are real, but they are still separate subsystems rather than one canonical dispute-ready evidence model for early external exchange.`
- Judgment: `stabilize`
- Execution track: `Stabilization Track`
- Secondary capability areas:
  - `Execution, Continuity & Accountability`
  - `External Collaboration & Economic Exchange`
- Secondary tracks:
  - `P2P Knowledge Exchange Track`

### Findings

1. `Major` Audit logging already records tool execution, policy decisions, sandbox decisions, and alerts.
   - This is a real accountability substrate rather than a placeholder.
   - References: `internal/observability/audit/recorder.go:23-30`, `internal/observability/audit/recorder.go:32-47`, `internal/observability/audit/recorder.go:50-71`, `internal/observability/audit/recorder.go:100-121`, `internal/observability/audit/recorder.go:124-140`

2. `Major` Provenance bundles already support redaction-aware export/import plus signature verification, including optional PQ dual signatures.
   - Bundles require signer DID plus signer implementation, verify against injected verifiers, and keep redaction as part of the exported bundle model.
   - References: `docs/features/provenance.md:11-16`, `docs/features/provenance.md:21-30`, `internal/provenance/bundle.go:15-31`, `internal/provenance/bundle.go:61-121`, `internal/provenance/bundle.go:124-188`

3. `Major` Provenance export is still wired to a wallet-backed v1 DID path, which narrows the otherwise broader signer/KMS story.
   - The P2P provenance exporter currently requires wallet-backed DID identity and explicitly uses the wallet v1 DID because the verification path only supports that shape.
   - At the same time, the broader security subsystem exposes multiple KMS/HSM providers as first-class crypto backends.
   - This means cryptographic accountability is real, but not yet one unified signer story.
   - References: `docs/features/provenance.md:69-75`, `internal/app/wiring_provenance.go:110-129`, `internal/security/kms_factory.go:9-44`, `internal/config/types_security.go:17-41`, `internal/config/types_security.go:128-133`

4. `Major` The product now has a lite dispute-ready receipt surface, but it is still not the full operator-facing evidence package.
   - The landed slice covers submission receipts, transaction receipts, current submission pointers, canonical current state, and append-only event trails.
   - Audit records, provenance bundles, and payment receipts still remain separate systems, so deeper provenance, settlement, and dispute integration are still pending.
   - The slice should not be described as a dispute engine or human dispute workflow.
   - References: `docs/security/dispute-ready-receipts.md:1-46`, `docs/architecture/p2p-knowledge-exchange-track.md:46-51`, `docs/features/provenance.md:7-16`, `internal/observability/audit/recorder.go:23-30`, `internal/provenance/bundle.go:161-188`

### Assessment

- `Auditability, Provenance & Cryptographic Accountability` is a core capability and should be kept.
- The correct action is `stabilize`:
  - keep the existing audit and provenance layers,
  - clarify how signer/KMS options relate to provenance export,
  - define one operator-facing evidence package for approval and dispute handling.

## Next Plan

The next follow-on work should stay narrow and directly build on this audit:

1. design an explicit `exportability policy` for `knowledge exchange v1`
2. design the `approval flow` for early external artifact exchange
3. complete deeper provenance, settlement, and dispute integration around the landed lite `dispute-ready receipts` model
