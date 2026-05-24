## ADDED Requirements

### Requirement: Dirty-repository workbench starter prompts mention changed targets when available

The standalone workbench SHALL mention the most obvious changed files or directories in the dirty-repository starter prompt when lightweight Git status output can be summarized.

#### Scenario: Dirty repository prompt highlights changed targets
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **AND** the changed targets can be summarized from lightweight Git status output
- **THEN** the dirty-repository starter prompt SHALL mention the current branch
- **AND** SHALL mention the summarized changed files or directories

#### Scenario: Dirty repository prompt falls back when changed targets are unclear
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a Git repository with uncommitted changes
- **AND** changed targets cannot be summarized
- **THEN** the dirty-repository starter prompt SHALL still mention the current branch and uncommitted changes
- **AND** SHALL NOT fail the startup path
