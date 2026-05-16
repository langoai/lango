# Tasks

## 1. Runtime Admission Config Surface

- [x] Add the runtime admission mode config surface with `off` and `observe`.
  Affected artifacts: `specs/cli-settings/spec.md`; runtime config schema and settings wiring when implementation starts.
  Verification: `openspec validate runtime-admission-boundary-hardening --strict`; config serialization/defaulting coverage confirms the new mode values.
- [x] Add fallback confidence defaults for the learning producer group and the librarian producer group without introducing extra first-slice producer groups.
  Affected artifacts: `specs/cli-settings/spec.md`; runtime config defaults when implementation starts.
  Verification: review confirms only the learning and librarian producer groups are named in this slice and that their defaults are fixed at `0.60` and `0.50`.

## 2. Graph Admission Classification Contract

- [x] Add the observe-only graph admission policy contract for the supported runtime graph inputs in this slice.
  Affected artifacts: `specs/graph-store/spec.md`; admission policy implementation and runtime wiring when implementation starts.
  Verification: strict OpenSpec validation passes; contract review confirms write routing remains observe-only.
- [x] Fix the observe-only admission decision taxonomy and counting units.
  Affected artifacts: `proposal.md`, `design.md`, `specs/graph-store/spec.md`, `specs/eventbus/spec.md`, `specs/system-feedback/spec.md`, `specs/cockpit-status-page/spec.md`.
  Verification: review confirms each observed slice is batch-scoped and carries `known_count`, `unknown_count`, and `unvalidated_count`, while baseline families retain their own batch/triple counting units.
- [x] Define graph-admission telemetry contracts around stable event-bus producer-source identifiers, event-bus producer-group identifiers, the synthetic `content_saved_extractor` telemetry source label, and validator-source tags.
  Affected artifacts: `specs/eventbus/spec.md`; observability event types when implementation starts.
  Verification: review confirms telemetry distinguishes admission classification from downstream write failures.
- [x] Define the aggregate graph write-failure baseline event contract as a separate downstream write-failure telemetry family.
  Affected artifacts: `specs/eventbus/spec.md`; observability event types when implementation starts.
  Verification: review confirms the aggregate write-failure baseline is not specified as carrying admission source, producer-group, or validator-source tags.

## 3. Supported Producer-Source Coverage

- [x] Observe `TriplesExtractedEvent` batches only for the explicit event-bus producer-source set `conversation_analysis`, `session_learning`, `learning`, and `proactive_librarian`.
  Affected artifacts: `specs/graph-store/spec.md`; runtime event-bus subscribers when implementation starts.
  Verification: review confirms no other event-bus source labels are normalized into named producer groups in this slice.
- [x] Observe `content.saved` extraction triples and the extractor dropped-unknown baseline under the separate synthetic `content_saved_extractor` telemetry source label without changing extractor behavior.
  Affected artifacts: `design.md`, `specs/graph-store/spec.md`, `specs/eventbus/spec.md`; extractor/graph wiring when implementation starts.
  Verification: review confirms dropped-unknown telemetry is source-scoped and separate from graph admission decisions.

## 4. Shared Validator Source And Observability Surfacing

- [x] Reuse the ontology predicate validator closure as the shared predicate-validity source for observe-only admission and graph-store validation.
  Affected artifacts: `specs/ontology-registry/spec.md`; ontology/graph integration when implementation starts.
  Verification: review confirms one validator source is named across both paths and that ontology init failure preserves existing graph validation behavior by degrading observe-only admission to an `unvalidated` observation mode.
- [x] Surface graph admission, dropped-unknown, unmapped-source, validator-source tag, and graph write-failure metrics in observability and cockpit status.
  Affected artifacts: `specs/eventbus/spec.md`, `specs/system-feedback/spec.md`, `specs/cockpit-status-page/spec.md`, proposal/design references in this change; observability and cockpit surfacing when implementation starts.
  Verification: review confirms event types carry the required identities; runtime feedback snapshots preserve the defined batch/triple counting units and grouped identities; `validator-source` remains a grouping key using the stable values `ontology_registry` and `unavailable` rather than a separate metric family; and the cockpit status page renders those grouped metrics.

## 5. Change Documentation And Verification

- [x] Keep proposal, design, and delta specs aligned on stable producer-source and config terminology.
  Affected artifacts: `proposal.md`, `design.md`, `specs/cli-settings/spec.md`, `specs/graph-store/spec.md`, `specs/ontology-registry/spec.md`.
  Verification: terminology review finds no relative scope wording that depends on transient wiring descriptions and no ambiguous fallback semantics for validator selection.
- [x] Validate the change scaffolding before implementation work begins.
  Affected artifacts: the full `openspec/changes/runtime-admission-boundary-hardening/**` change set.
  Verification: run `openspec validate runtime-admission-boundary-hardening --strict`.
