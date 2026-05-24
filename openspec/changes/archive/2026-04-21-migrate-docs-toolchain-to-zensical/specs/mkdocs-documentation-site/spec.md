## MODIFIED Requirements

### Requirement: Zensical configuration file
The project SHALL have a `zensical.toml` at the repository root configuring the canonical docs site with public navigation, search, dark/light mode, code copy support, and Mermaid rendering.

#### Scenario: Valid Zensical configuration
- **WHEN** `.venv/bin/zensical build` is run from the project root
- **THEN** the site builds successfully and produces a static site in `site/`

#### Scenario: Theme features enabled
- **WHEN** the documentation site is served
- **THEN** it SHALL have navigation tabs, search suggestions, code copy buttons, dark/light mode toggle, and Mermaid diagram rendering

### Requirement: Documentation directory structure
The `docs/` directory SHALL contain public markdown files organized into subdirectories: `getting-started/`, `architecture/`, `features/`, `automation/`, `security/`, `payments/`, `cli/`, `gateway/`, `deployment/`, `development/`, and root-level `index.md` and `configuration.md`. Hidden support docs and withdrawn cockpit sub-guides SHALL live outside `docs/`.

#### Scenario: All navigation entries resolve
- **WHEN** the docs site is built
- **THEN** every entry in the public navigation SHALL resolve to an existing public markdown file

### Requirement: Hidden docs are outside the public docs tree
The project SHALL keep hidden docs, superpowers planning artifacts, and withdrawn cockpit sub-guides out of `docs/` so the public site is represented structurally rather than through exclusion rules.

#### Scenario: Hidden docs do not ship in the public site
- **WHEN** `.venv/bin/zensical build` is run from the project root
- **THEN** the hidden files SHALL not appear in the generated site or public navigation

### Requirement: Public site navigation exposes the chosen public surfaces
The docs site navigation SHALL surface the selected public Security pages and the top-level Research page while keeping the rest of the documentation tree hidden from the public site.

#### Scenario: Public security and research pages remain navigable
- **WHEN** the documentation site is built
- **THEN** the nav SHALL include only the intended public Security entries and the Research entry
