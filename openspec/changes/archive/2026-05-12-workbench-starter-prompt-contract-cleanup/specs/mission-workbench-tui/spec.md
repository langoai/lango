## ADDED Requirements

### Requirement: Workbench starter prompts share one generation contract

The standalone workbench SHALL derive starter prompt defaults and starter prompt context-shaping from one shared generation contract rather than from duplicated page-local fallback strings.

#### Scenario: Shared default prompts backstop the workbench page
- **WHEN** the Mission Control workbench page needs its fallback starter prompts
- **THEN** it SHALL use the same shared starter prompt contract that the workbench shell uses to build context-aware prompt sets
- **AND** quick-start copy SHALL remain behaviorally consistent across the workbench shell and page rendering path
