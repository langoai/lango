# MkDocs IA And Strict Baseline Recovery Design

## Purpose

This design defines a documentation-structure cleanup slice for the MkDocs site.

The immediate goals are:

- restore a clean `python3 -m mkdocs build --strict` baseline
- make the public documentation site reflect a deliberate information architecture instead of exposing or warning on every file under `docs/`

This is not a full documentation rewrite. It is a focused information-architecture pass that clarifies what should be publicly surfaced, what should remain internal, and how MkDocs should treat both categories.

## Scope

This slice covers:

- MkDocs navigation cleanup in `mkdocs.yml`
- explicit inclusion of selected user/operator-facing documents
- explicit exclusion of internal or work-in-progress documents from the built site
- cockpit sub-guide consolidation into the main cockpit entry document
- repair of the broken installation anchor link in the getting-started flow
- restoration of a passing `mkdocs --strict` build

This slice does not cover:

- rewriting the entire docs site structure
- large-scale content refresh across all product docs
- moving or deleting broad sets of source files outside the docs tree
- publishing internal design worklogs or implementation plans as end-user docs

## Information Architecture Principles

The documentation site should only surface documents that are intentionally published for readers of the built site.

This slice adopts five rules:

1. `nav` contains only documents that should be publicly discoverable from the site.
2. Operationally important deep-dive docs may remain public under their existing domain section.
3. Research documents are separated from product and operator docs under a dedicated `Research` section.
4. Internal design artifacts and worklogs are not part of the public site.
5. Multi-file sub-guides that fragment one feature area should be consolidated into a single primary entry where practical.

## Public vs Hidden Content

### Public

The site will continue to expose the current product/operator docs already in `nav`, plus the following additions:

- `security/approval-cli.md`
- `security/envelope-migration.md`
- `research/phase7-pq-onchain-feasibility.md`

### Hidden

The following content should not be part of the built public site:

- `superpowers/specs/*`
- `superpowers/plans/*`
- `architecture/adr-001-package-boundaries.md`
- `architecture/dependency-graph.md`
- cockpit sub-guides:
  - `features/cockpit-approval-guide.md`
  - `features/cockpit-channels-guide.md`
  - `features/cockpit-tasks-guide.md`
  - `features/cockpit-troubleshooting.md`

The hidden set is not merely removed from `nav`. It must be excluded from the MkDocs site build so `--strict` no longer warns about these files being present but unlisted.

## Navigation Changes

### Security

The `Security` section should explicitly expose the following additional docs:

- `Approval CLI`
- `Envelope Migration`

These are not internal notes. They are operator/developer-facing security docs and belong with the rest of the security surface.

### Research

A new top-level `Research` section should be added with:

- `research/phase7-pq-onchain-feasibility.md`

This keeps research and exploratory material visible without mixing it into product architecture or operator guidance.

### Cockpit

`features/cockpit.md` remains the single public cockpit entry document.

The detailed cockpit sub-guides are not directly published as separate site pages in this slice. Instead, their useful content should be absorbed or summarized in the primary cockpit document as needed, while the source files themselves are excluded from the site build.

## Build Exclusion Strategy

This slice requires explicit exclusion of hidden documents from the MkDocs build target.

The implementation should prefer a clear, maintainable exclusion approach in `mkdocs.yml` or the site-build configuration rather than a workaround that merely suppresses warnings.

The intended result is:

- public docs remain buildable and searchable
- hidden docs remain in the repository
- hidden docs do not produce strict-mode warnings
- hidden docs are not reachable as built site pages

## Broken Link Recovery

The current strict build also reports a broken anchor link in:

- `docs/getting-started/quickstart.md`

It links to:

- `installation.md#platform-specific-c-compiler-setup`

but the current installation page exposes:

- `#optional-c-compiler-setup`

This link must be corrected as part of the slice so the getting-started path is internally consistent again.

## Recommended Approach

Three approaches were considered:

### 1. Nav-Only Cleanup

Add more pages to `nav` and leave the rest in the docs tree.

Pros:

- fast

Cons:

- keeps internal documents implicitly public
- does not satisfy the “hide and exclude” requirement
- tends to bloat navigation with material that should not be public

### 2. Nav Plus Explicit Exclusion

Deliberately choose which docs belong in the public site, add those to `nav`, and explicitly exclude the hidden set from the site build.

Pros:

- directly restores strict-mode health
- matches the chosen IA model
- preserves internal docs in-repo without exposing them publicly
- keeps public navigation intentional and readable

Cons:

- slightly more involved than a nav-only change

### 3. Large-Scale Docs Reorganization

Move files around broadly or redesign the entire docs tree.

Pros:

- could produce the cleanest long-term structure

Cons:

- too large for this slice
- creates unnecessary churn

Recommended choice: **Approach 2: Nav Plus Explicit Exclusion**.

## Concrete Changes

This slice should make the following concrete changes:

1. Update `mkdocs.yml`
   - add `Security` entries for `approval-cli.md` and `envelope-migration.md`
   - add a new top-level `Research` section
   - configure the site build so hidden documents are excluded

2. Update `features/cockpit.md`
   - absorb or summarize the important operator-facing guidance from the cockpit sub-guides
   - keep cockpit as one public entry page

3. Fix the broken installation anchor link
   - update `quickstart.md` to point at the actual installation anchor

4. Preserve hidden docs in-repo
   - no need to delete the underlying markdown files
   - they simply stop participating in the public site build

## Validation

This slice is successful only if all of the following are true:

- `python3 -m mkdocs build --strict` passes
- the new `Security` and `Research` nav entries appear as intended
- hidden docs no longer appear as “present but not included in nav”
- the broken installation anchor warning is gone
- cockpit guidance remains accessible through the main cockpit page without exposing the sub-guides individually

## Follow-On Work

This slice deliberately stops once the MkDocs strict baseline and public IA are restored.

Possible later follow-on work includes:

- deeper cockpit content consolidation
- broader doc style normalization
- a full review of which architecture deep dives should eventually be published or hidden
- a separate policy for publishing or not publishing internal design records over time
