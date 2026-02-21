## Why

The Lango project has a comprehensive 907-line monolithic README.md but no dedicated documentation site. A MkDocs Material documentation site provides professional, searchable, navigable documentation with dark/light mode, making the project more accessible and easier to maintain.

## What Changes

- Add `mkdocs.yml` configuration at project root with MkDocs Material theme, search, code copy, mermaid diagrams, and dark/light mode toggle
- Add `docs/` directory with 47 markdown pages organized into 11 sections (Getting Started, Architecture, Features, Automation, Security, Payments, CLI Reference, Gateway & API, Deployment, Development, Configuration Reference)
- Add `docs/assets/logo.png` (copy of project logo) and `docs/stylesheets/extra.css` for status badges
- Content sourced from README.md sections and source code analysis, broken into focused navigable pages with Mermaid architecture diagrams, tables, admonitions, and cross-links

## Capabilities

### New Capabilities

- `mkdocs-documentation-site`: MkDocs Material documentation site with 47 markdown pages, custom CSS, navigation structure, and all project features documented

### Modified Capabilities

(none — this is a documentation-only addition with no code changes)

## Impact

- No Go code changes — build and tests unaffected
- New files: `mkdocs.yml`, `docs/` directory (49 files total)
- New dev dependency: `mkdocs-material` and `mkdocs-minify-plugin` (Python, not added to Go modules)
- Deployable to GitHub Pages via `mkdocs gh-deploy`
