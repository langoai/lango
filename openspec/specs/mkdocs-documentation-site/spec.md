# mkdocs-documentation-site Specification

## Purpose
Zensical-native documentation site with an explicit public IA and feature parity for the public docs experience.
## Requirements
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

### Requirement: Home page with feature grid
The `docs/index.md` SHALL display a feature grid using Material grid cards, an experimental warning admonition, a quick install snippet, and links to getting started.

#### Scenario: Home page renders
- **WHEN** a user visits the documentation site root
- **THEN** they SHALL see the project name, feature grid cards, and navigation to all sections

### Requirement: Getting Started section
The documentation SHALL include installation prerequisites (Go 1.25+, CGO), build instructions, quick start guide with onboard wizard walkthrough, and configuration basics.

#### Scenario: New user onboarding path
- **WHEN** a new user follows the Getting Started section
- **THEN** they SHALL find step-by-step instructions from installation through first run

### Requirement: Architecture documentation with diagrams
The documentation SHALL include a system overview with Mermaid architecture diagram, project structure with package descriptions, and data flow with Mermaid sequence diagram.

#### Scenario: Architecture diagrams render
- **WHEN** the architecture pages are viewed
- **THEN** Mermaid diagrams SHALL render showing system layers and data flow

### Requirement: Feature documentation coverage
The documentation SHALL have dedicated pages for: AI Providers, Channels, Knowledge System, Observational Memory, Embedding & RAG, Knowledge Graph, Multi-Agent Orchestration, A2A Protocol, P2P Network, P2P Economy, Smart Contracts, Observability, Skill System, Proactive Librarian, and System Prompts.

#### Scenario: All features documented
- **WHEN** a user browses the Features section
- **THEN** each feature SHALL have its own page with configuration reference and usage examples

### Requirement: CLI reference documentation
The documentation SHALL include a complete CLI reference organized by command category: Core, Config Management, Agent & Memory, Security, Payment, P2P, Economy, Contract, Metrics, and Automation commands.

#### Scenario: CLI commands documented
- **WHEN** a user looks up a CLI command
- **THEN** they SHALL find syntax, flags, and usage examples

### Requirement: Navigation includes P2P pages
The site navigation SHALL include "P2P Network: features/p2p-network.md", "P2P Economy: features/economy.md", "Smart Contracts: features/contracts.md", and "Observability: features/observability.md" in the Features section and "P2P Commands: cli/p2p.md", "Economy Commands: cli/economy.md", "Contract Commands: cli/contract.md", and "Metrics Commands: cli/metrics.md" in the CLI Reference section.

#### Scenario: P2P feature in nav
- **WHEN** the docs site is built
- **THEN** the Features navigation section includes a "P2P Network" entry after "A2A Protocol"

#### Scenario: Economy, contract, observability features in nav
- **WHEN** the docs site is built
- **THEN** the Features navigation section includes "P2P Economy", "Smart Contracts", and "Observability" entries after "P2P Network"

#### Scenario: P2P CLI in nav
- **WHEN** the docs site is built
- **THEN** the CLI Reference navigation section includes a "P2P Commands" entry after "Payment Commands"

#### Scenario: Economy, contract, metrics CLI in nav
- **WHEN** the docs site is built
- **THEN** the CLI Reference navigation section includes "Economy Commands", "Contract Commands", and "Metrics Commands" entries after "P2P Commands"

### Requirement: Public site navigation exposes the chosen public surfaces
The docs site navigation SHALL surface the selected public Security pages and the top-level Research page while keeping the rest of the documentation tree hidden from the public site.

#### Scenario: Public security and research pages remain navigable
- **WHEN** the documentation site is built
- **THEN** the nav SHALL include only the intended public Security entries and the Research entry

### Requirement: Configuration reference
The documentation SHALL include a complete configuration reference page listing all configuration keys with type, default value, and description, organized by category, including Economy and Observability sections.

#### Scenario: Configuration completeness
- **WHEN** the configuration reference is viewed
- **THEN** it SHALL list all configuration keys including economy.* and observability.* sections

### Requirement: Assets and custom CSS
The `docs/assets/` directory SHALL contain the project logo. The `docs/stylesheets/extra.css` SHALL define badge styles for experimental and stable feature status indicators.

#### Scenario: Logo and styles present
- **WHEN** the documentation site is served
- **THEN** the logo SHALL appear in the navigation header and badge CSS SHALL be available
