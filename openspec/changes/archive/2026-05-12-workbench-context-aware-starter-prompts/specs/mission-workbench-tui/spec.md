## ADDED Requirements

### Requirement: Ready-profile workbench starter prompts reflect workspace context

The standalone workbench SHALL adapt its ready-profile starter prompts to the detected workspace context when `lango` starts inside a project.

#### Scenario: Repository-aware prompts in a detected project
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workdir belongs to a repository
- **THEN** the starter prompts SHALL reference the detected repository instead of using only generic copy

#### Scenario: Go-aware structure prompt in a Go module
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workdir belongs to a workspace containing a `go.mod`
- **THEN** the structure-oriented starter prompt SHALL use Go package layout guidance instead of a generic project-structure prompt

#### Scenario: Generic fallback outside a detected project
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** no repository markers are detected from the current workdir
- **THEN** the starter prompts SHALL fall back to the generic repository-summary, project-structure, and recent-change prompts
