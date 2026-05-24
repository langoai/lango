## Context

The site already has the content needed for the public documentation surface, but the current MkDocs configuration and docs entry points do not cleanly separate public pages from hidden support material. That creates strict-build noise and makes the public IA harder to trust.

## Goals / Non-Goals

**Goals:**

- Keep hidden documentation out of the public build.
- Surface only intentionally published Security and Research pages.
- Preserve the existing public docs content while fixing the broken quickstart anchor.
- Reduce cockpit entry-point fragmentation by consolidating operator guidance.

**Non-Goals:**

- Rewrite unrelated documentation sections.
- Expand the public security surface beyond the selected recovery slice.
- Rework the cockpit feature beyond the documentation consolidation described in this slice.

## Decisions

### Decision: Use exclusion-aware MkDocs IA

The MkDocs site should treat hidden docs as non-public build inputs rather than trying to document or nav-link every file in the repository.

Why:
- This restores the strict build baseline without exposing internal support content.
- It keeps the public IA focused on intentionally published docs.

### Decision: Keep security exposure selective

The security documentation surface should expose the selected deep-dive docs that belong in the public IA and leave the rest of the security tree unchanged.

Why:
- The slice is about truth-aligned exposure, not a full security-site redesign.

### Decision: Consolidate cockpit public entry points

The main cockpit page should become the public entry point for operator guidance that was previously split across separate public sub-guides.

Why:
- The consolidated page is easier to navigate and matches the narrowed public IA.

## Risks / Trade-offs

- A smaller public nav may look less comprehensive, but that is intentional.
- Consolidating cockpit guidance increases the size of the main cockpit page, but reduces fragmentation.
- The quickstart anchor fix is tiny, but it removes a strict-build failure and a broken user path.
