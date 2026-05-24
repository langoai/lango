## Context

The repository has already moved most advanced configuration guidance to `lango settings`, but a few stale references remained in README and feature docs. Those stale references are especially harmful because they suggest product surfaces that no longer exist: a broader `lango onboard` scope and a plaintext `config.yaml`.

## Goals / Non-Goals

**Goals:**
- Remove the remaining false configuration-path references from public docs.
- Reassert encrypted-profile storage as the canonical configuration model.
- Keep the slice documentation-only and low risk.

**Non-Goals:**
- Changing onboarding scope.
- Altering runtime configuration behavior.
- Rewriting the broader docs IA.

## Decisions

### D1: Prefer `lango settings` for advanced interactive configuration
Where a doc discusses advanced feature configuration, it now points to `lango settings` first and to `lango config import/export` when a programmatic flow is relevant.

### D2: Remove `config.yaml` phrasing rather than mapping it to a file path
The product no longer uses a plaintext config file as the canonical source of truth, so the docs now talk about config structure managed through the settings editor or import/export workflow.

## Risks / Trade-offs

- **[Docs become less file-oriented]** → Some users may expect a direct file to edit. Mitigation: the docs still show JSON structure for import/export while avoiding false claims about a canonical local config file.
