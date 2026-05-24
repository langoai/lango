# Zensical Native Migration Design

## Purpose

This design defines the migration of the documentation toolchain away from MkDocs and Material for MkDocs to Zensical.

The problem is no longer just a warning during docs builds. The current toolchain is structurally misaligned with the direction of the upstream ecosystem:

- MkDocs 1.x is effectively stalled
- MkDocs 2.0 is not compatible with Material for MkDocs
- Material for MkDocs now explicitly warns users to plan for a different long-term path
- Zensical is presented by the Material team as the intended long-term migration path for existing MkDocs 1.x-style sites

This slice therefore treats the issue as a toolchain migration, not a warning suppression exercise.

## Scope

This slice covers:

- replacing the current MkDocs-based site configuration with Zensical-native site configuration
- redefining the public documentation site structure under the new toolchain
- migrating the docs build workflow used locally and in CI
- preserving the current public documentation experience where it matters most:
  - site structure
  - link integrity
  - search
  - dark mode
  - code copy
  - Mermaid rendering

This slice does not cover:

- a full rewrite of the documentation content
- a broad visual redesign of the docs site
- a large-scale restructuring of the `docs/` source tree
- production deployment verification beyond the configuration and workflow handoff

The objective is a working Zensical-native documentation stack with feature continuity, not a full documentation rebrand.

## Recommended Approach

Three broad approaches were considered:

### 1. Tool Swap Only

Keep the existing docs structure and mostly try to translate the current MkDocs configuration mechanically.

Pros:

- faster initial change

Cons:

- leaves too much MkDocs-era structure and assumption behind
- tends to create a fragile compatibility layer instead of a clean native configuration

### 2. Native Migration

Replace the canonical site configuration, keep the source docs where practical, and rebuild the public documentation contract natively around Zensical.

Pros:

- matches the actual long-term goal
- avoids carrying forward obsolete MkDocs-only assumptions
- lets the project define a clean, durable documentation toolchain

Cons:

- larger than a simple tool swap

### 3. Full Site Redesign

Use the migration as a trigger to redesign IA, theme, content structure, and presentation all at once.

Pros:

- maximum long-term flexibility

Cons:

- far too large for a first migration slice

Recommended choice: **Approach 2: Native Migration**.

## Migration Strategy

The migration should not be treated as a one-to-one configuration translation.

Instead:

1. `mkdocs.yml` stops being the canonical site definition.
2. A Zensical-native site configuration becomes the single source of truth.
3. The `docs/` source tree remains largely intact unless a specific source-level move is justified.
4. CI and local build entrypoints are migrated in the same slice.
5. Compatibility is measured by user-facing behavior, not by preserving old configuration syntax.

This means the work is conceptually:

- re-declare the public documentation site in Zensical
- keep the public documentation experience stable
- remove the project’s dependency on the MkDocs/Material stack

## Compatibility Contract

The migration must preserve the following user-facing guarantees as closely as practical:

- the existing public documentation structure remains recognizable
- the current public sections remain available:
  - Home
  - Getting Started
  - Architecture
  - Features
  - Automation
  - Security
  - Research
  - Payments
  - CLI Reference
  - Gateway & API
  - Deployment
  - Development
  - Configuration Reference
- existing public internal links continue to resolve
- Mermaid diagrams still render
- code blocks still offer copy support
- dark mode still exists
- search remains available

The migration does **not** need to preserve:

- Material-specific settings
- Material-specific visual details
- `mkdocs.yml` compatibility
- plugin-specific implementation details that are invisible to readers

The compatibility target is the public documentation experience, not configuration syntax.

## Concrete Outputs

This slice should produce the following outputs:

### 1. Zensical Site Configuration

A new canonical site configuration for Zensical that defines:

- public navigation
- docs source integration
- rendering behavior
- feature toggles needed for:
  - search
  - dark mode
  - code copy
  - Mermaid

### 2. Local Docs Build Contract

A clear local workflow for:

- building the docs
- serving or previewing the docs

The project should no longer depend on `mkdocs build` as the canonical local docs command after the migration.

### 3. CI Docs Workflow Migration

The GitHub Actions docs workflow should be updated so it no longer installs and builds through the old MkDocs/Material stack.

It should use the new Zensical-native build path instead.

### 4. Repository Docs References Update

Any repository-level references that still describe the docs toolchain as MkDocs-based should be updated.

This includes places like:

- README
- development docs
- architecture/project-structure docs
- docs build references in workflow or operator documentation

Only references that are actually affected by the migration need to change.

### 5. Removal Of Old Canonical Tooling

The migration should remove or retire:

- `mkdocs.yml` as the canonical site config
- workflow steps that install or invoke MkDocs/Material for the docs site
- project-level assumptions that the docs site is built through MkDocs

## Validation

This slice is successful only if all of the following are true:

- the documentation site builds successfully through the new Zensical-native path
- CI docs workflow is updated to the same path
- the public docs structure still reflects the current public IA
- internal links remain valid
- search, dark mode, code copy, and Mermaid are still present
- the project no longer depends on the old MkDocs/Material flow as its canonical documentation build

## Follow-On Work

This slice deliberately stops short of broader site redesign work.

Possible later follow-on work includes:

- deployment validation for the new docs toolchain in GitHub Pages or the chosen host
- visual refinement of the migrated site
- a second-pass cleanup of documentation structure once the migration is stable
- further documentation authoring rules tailored to Zensical
