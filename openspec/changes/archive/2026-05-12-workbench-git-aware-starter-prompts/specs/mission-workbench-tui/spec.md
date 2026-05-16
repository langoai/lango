## ADDED Requirements

### Requirement: Ready-profile workbench starter prompts reflect live Git state when available

The standalone workbench SHALL use lightweight Git signals to sharpen the ready-profile change-review starter prompt when the detected workspace is a Git repository and Git metadata is available.

#### Scenario: Clean repository prompt mentions the current branch
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with a current branch and no uncommitted changes
- **THEN** the change-review starter prompt SHALL mention the current branch

#### Scenario: Dirty repository prompt mentions uncommitted changes
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **THEN** the change-review starter prompt SHALL mention the uncommitted changes and the current branch

#### Scenario: Git failure keeps the repository-aware fallback
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is repository-shaped but Git metadata cannot be read
- **THEN** the change-review starter prompt SHALL fall back to the repository-aware non-Git wording instead of failing the workbench startup path
