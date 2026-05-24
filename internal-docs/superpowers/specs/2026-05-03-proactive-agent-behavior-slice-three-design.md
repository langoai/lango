# Proactive Agent Behavior Slice 3 Design

## Strategic Goal

Slice 2 made missions durable and made Mission Control read durable mission rows first. Slice 3 adds proactive behavior on top of that foundation: agents should be able to notice useful work before the user asks explicitly, prepare low-risk work automatically, and surface that work as actionable proposed missions.

The key constraint is control. Slice 3 should make Lango feel more agentic without creating a system that silently performs risky work or mutates durable mission truth without a clear user-facing reason.

## Product Intent

Slice 3 should answer a simple user question:

> "What useful work did Lango notice, what has it already prepared for me, and what still needs my decision?"

The user should see:

- proposed missions created from observed context
- what preparation work has already happened
- what evidence or draft artifacts are ready
- what would happen next if the user accepts or rejects the proposal

## Fixed Decisions

The following decisions are fixed for Slice 3:

- `proposed` remains a transient proposal state, not a durable `Mission` row
- durable mission rows still begin at `prepared` or `active`
- agents may automatically perform only low-risk preparation work
- risky, irreversible, external, expensive, or state-mutating actions still require a live decision
- proposal preparation must remain explainable and attributable to a visible proposal
- proposal acceptance must convert visible prepared work into a durable mission path instead of discarding it

## Why `proposed` Stays Transient

Slice 2 deliberately excluded durable `proposed` mission rows. That was the right tradeoff and should remain true in Slice 3.

Reasons:

- user-started and accepted work remain the durable mission truth
- many proactive suggestions are ephemeral and should expire quietly
- forcing every suggestion into the mission table would pollute the main durable work graph
- proposal-specific preparation needs a different lifecycle than accepted mission work

Slice 3 therefore needs a dedicated transient proposal model rather than overloading durable missions.

## Proposal Model

Slice 3 introduces a transient `Proposal` concept.

Suggested fields:

- `proposal_id`
- `session_key`
- `source_kind`
- `source_ref`
- `title`
- `summary`
- `reason`
- `confidence`
- `status`
- `created_at`
- `updated_at`
- `expires_at`

Suggested statuses:

- `suggested`
- `preparing`
- `prepared`
- `dismissed`
- `accepted`
- `expired`

This does not need to be durable in the same way as `Mission` in the first Slice 3 slice. A session-scoped or short-lived persisted registry is enough, as long as Mission Control can render it honestly and expiration/dismissal are deterministic.

## Proposal Producers

Slice 3 should start from producers that already exist in the repo or are close enough to current behavior to be credible.

Initial proposal producers:

1. `LearningSuggestionEvent`
2. proactive librarian gaps and inquiries
3. repeated runtime failures or blocked work with a clear next step

The first slice should avoid broad, vague heuristics such as "guess what the user wants today" unless there is a concrete producer.

## Producer Contracts

Slice 3 should not treat all "possible future producers" as ready on day one. Each producer needs an explicit contract.

### Producer 1: Learning Suggestions

This is the only producer that already reaches Mission Control today in a proposal-like form.

Phase 1 contract:

- input: `LearningSuggestionEvent`
- output: one transient proposal candidate
- source identity: `source_kind=proposed_learning`, `source_ref=SuggestionID`
- preparation allowed: yes, subject to proposal policy

### Producer 2: Proactive Librarian

The librarian has real gap/inquiry machinery, but it does not yet emit proposal records.

Phase 1 contract:

- input: a dedicated adapter over pending gaps/inquiries, not ad hoc UI scraping
- output: transient proposal candidate for research/brief preparation
- source identity: stable inquiry or gap key
- gating: do not enable until the adapter exists explicitly

In other words, librarian is an intended Slice 3 producer, but not an implicit one.

### Producer 3: Runtime Failure / Blocked Work

Runtime failures and blocked work already exist, but they are not yet proposal candidates by themselves.

Phase 1 contract:

- input: explicit adapter over stable runtime failure or blocked-state signals
- output: transient proposal candidate such as retry plan, prerequisite gathering, or follow-up brief
- source identity: stable execution or mission reference plus failure class
- gating: do not enable until the adapter exists explicitly

This prevents the first implementation from drifting into vague heuristics.

### Learning Suggestions

These already exist and already surface as transient proposed missions. Slice 3 should promote them from passive suggestions into proposal records that can own preparation work.

### Proactive Librarian

The librarian already detects gaps, creates inquiries, and can auto-save high-confidence knowledge. Slice 3 should let those gaps create proposals such as:

- research missing context
- summarize unresolved questions
- prepare a short brief before asking the user

### Runtime Failure Recovery

When runtime work repeatedly fails or blocks with a stable reason, Slice 3 may generate a proposal such as:

- prepare a retry plan
- gather missing prerequisites
- draft the next manual action

This should be deterministic and conservative.

## Preparation Policy

Slice 3 must define what agents may do automatically before acceptance.

Allowed automatic preparation categories:

- read
- search
- analyze
- summarize
- classify
- compare
- draft
- collect candidate options

Forbidden without live decision:

- filesystem mutation in the user workspace
- message sending
- calendar confirmation
- payment or budget spend
- contract writes
- dangerous command execution
- irreversible external side effects

If draft artifacts require storage, they should be written into a dedicated transient draft/preparation store, not directly into the user workspace.

## Preparation Execution

Proposal preparation needs its own execution identity.

Slice 3 should introduce a transient proposal execution relationship:

- a proposal can launch zero or more preparation jobs
- those jobs may use background execution
- those jobs must remain attributable to one visible proposal

This implies a `proposal_id` context binding for preparation work.

Suggested execution flow:

1. producer emits a candidate
2. proposal coordinator deduplicates and creates/updates a proposal
3. proposal policy decides whether preparation is allowed
4. preparation job executes low-risk work
5. Mission Control shows proposal status and prepared outputs

## Proposal Acceptance

When the user accepts a proposal:

1. create durable mission row through `MissionService.AcceptProposal(...)`
2. promote any proposal-owned prepared execution refs or draft artifacts that should survive acceptance
3. mark the proposal as accepted and remove it from the transient overlay
4. keep the durable mission as the long-term owner from that point onward

The important constraint is that acceptance should not throw away useful prework.

### Ownership Transfer Rule

Slice 3 must stay compatible with the Slice 2 rule that `MissionExecutionLink` is written at execution creation sites for durable missions.

That means:

- proposal preparation jobs are proposal-owned, not mission-owned
- proposal preparation must not write `MissionExecutionLink` rows before a durable mission exists
- the proposal registry stores transient preparation execution refs under `proposal_id`
- acceptance performs a one-time promotion step: attach proposal-owned execution refs to the newly created durable mission if they are not already linked elsewhere
- Slice 3 does not support arbitrary relinking of an execution ref that is already durably linked to another mission

This keeps Slice 2 durable link ownership rules intact.

## Mission Control Surface

Slice 3 should extend the existing proposed-mission surface rather than replacing it.

Each proposed mission should reveal:

- why it exists
- what source produced it
- whether preparation is running or ready
- what evidence or draft output is available
- what accepting it will do

Suggested compact fields:

- title
- one-line reason
- preparation state (`suggested`, `preparing`, `prepared`)
- prepared artifact count or summary
- updated time

Mission Control should still prioritize durable missions first. Proposed missions remain a secondary but visible lane.

## Approval and Decision Semantics

Slice 3 must preserve the live-decision discipline established earlier.

Rules:

- proposal preparation itself should avoid hitting approval whenever possible by staying inside low-risk work
- if preparation reaches a risky boundary, it must stop and surface a live decision
- acceptance of a proposal is not itself permission to perform unrelated risky actions

This keeps the system proactive but legible.

## Architecture Sketch

Suggested Slice 3 components:

- `ProposalCoordinator`
  - subscribes to proposal producers
  - deduplicates and updates proposals

- `ProposalPolicy`
  - decides whether to create, suppress, prepare, or expire proposals

- `ProposalRegistry`
  - stores active transient proposals plus their preparation metadata

- `ProposalPreparationRunner`
  - launches low-risk preparation work with `proposal_id` binding

The first implementation should keep these components narrow and avoid creating a second mission-sized domain stack.

## Relationship to Ontology

Slice 3 should remain compatible with the longer-term ontology direction without forcing ontology to become the runtime control plane immediately.

Near-term stance:

- proposals may eventually map to ontology concepts such as open loop, inquiry, or candidate goal
- durable missions remain the runtime work unit for accepted work
- Slice 3 should not require ontology writes for every proposal before the proposal model is stable

## Non-Goals

Slice 3 does not aim to:

- make all suggestions durable mission rows
- auto-approve risky work
- mutate user files silently
- replace live decision surfaces
- solve full work/life agenda synthesis
- introduce full multi-agent coworking visibility yet

Those belong to later slices.

## Canonical Note

This document supersedes earlier roadmap language that loosely listed `proposed` inside durable mission lifecycle states. The current canonical direction is:

- `proposed` stays transient
- durable mission rows still begin at `prepared` or `active`

## Success Criteria

Slice 3 is successful when:

- agents can produce visible proposed missions from real producers
- low-risk preparation work can happen automatically and be shown in Mission Control
- proposal acceptance converts into durable mission lifecycle without losing prepared context
- risky boundaries still stop at live decisions
- proposed missions remain transient and do not pollute durable mission truth
