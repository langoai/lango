## Context

Lango has a 907-line monolithic README.md covering all features, configuration, CLI reference, architecture, and deployment. As the project grows, a single README becomes harder to navigate and maintain. MkDocs Material provides a professional documentation site with search, navigation tabs, dark/light mode, and Mermaid diagram support.

## Goals / Non-Goals

**Goals:**

- Break README content into ~47 focused, navigable documentation pages
- Provide searchable documentation with MkDocs Material theme
- Include Mermaid architecture and data flow diagrams
- Support dark/light mode toggle
- Enable deployment to GitHub Pages via `mkdocs gh-deploy`
- Maintain accurate cross-links between related documentation pages

**Non-Goals:**

- Auto-generated API documentation from Go source code
- CI/CD pipeline for automatic documentation deployment
- Versioned documentation (multi-version support)
- Custom MkDocs plugins beyond search and minify
- Replacing README.md (it remains as a quick reference)

## Decisions

**MkDocs Material over alternatives (Hugo, Docusaurus, GitBook)**

MkDocs Material is the standard for developer documentation. It provides built-in search, code copy buttons, Mermaid support, admonitions, and dark/light mode without custom JavaScript. Python-based, which keeps it separate from the Go build pipeline. Alternatives require more configuration or have heavier runtime dependencies.

**Navigation structure with 11 top-level tabs**

Content organized into: Getting Started, Architecture, Features, Automation, Security, Payments, CLI Reference, Gateway & API, Deployment, Development, and Configuration Reference. This mirrors how users approach the project — from onboarding through feature exploration to reference material.

**Content sourced from README + source code analysis**

README provides the primary content. Architecture diagrams and data flow pages are derived from source code analysis of `internal/app/wiring.go`, `internal/app/app.go`, and package relationships. This ensures documentation accuracy.

**Minimal custom CSS**

Only `extra.css` for experimental/stable status badges. All other styling uses Material theme defaults. This minimizes maintenance burden.

## Risks / Trade-offs

[Documentation drift] Documentation may become outdated as code evolves → Keep README as quick reference; documentation site provides depth. Both should be updated together during feature changes.

[Python dependency] MkDocs requires Python and pip, adding a non-Go dependency for docs → Only needed for docs building, not for the Go application. Standard in the ecosystem.

[Large initial file count] 49 new files in a single change → Files are all documentation (no code changes), organized in clear directory structure. No impact on build or test.
