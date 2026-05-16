## Context

The workbench already has state-aware empty-state and composer guidance. The remaining inconsistency was the header summary, which could still show a provider label even when there was no usable provider config or model setup behind it.

## Goals / Non-Goals

**Goals:**
- Keep the workbench header honest about whether the active profile is actually ready.
- Reuse the same readiness heuristic already used by the workbench empty-state guidance.
- Preserve the richer provider/model summary for ready profiles.

**Non-Goals:**
- Redesigning the Mission Control header layout.
- Changing provider resolution or validation semantics.
- Altering cockpit-only behavior beyond the shared projector output.

## Decisions

### D1: Reuse the existing setup-readiness heuristic
The projector now reuses the same basic conditions already implied by the workbench guidance: provider ID present, model present, provider config present, provider type present, and non-Ollama API credentials available.

### D2: Use a plain `Setup required` label
The header space is small. A short, explicit label communicates the state better than trying to embed next-step instructions into the header itself.

## Risks / Trade-offs

- **[Heuristic remains conservative]** → Some profiles may still be functionally broken for reasons beyond provider/model setup. Mitigation: the header only claims readiness at a coarse level; `lango doctor` remains the deeper verifier.
- **[Shared projector affects cockpit too]** → The summary comes from shared projector logic. Mitigation: the change only affects clearly incomplete profiles and improves truthfulness on both surfaces.
