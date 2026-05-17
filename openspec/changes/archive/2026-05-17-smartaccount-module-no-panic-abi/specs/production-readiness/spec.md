## ADDED Requirements

### Requirement: Smart account module encoder avoids production panic paths

The `internal/smartaccount/module` package SHALL NOT contain production `panic` calls for deterministic ABI argument initialization or module calldata encoding.

#### Scenario: Smart account module ABI setup returns errors instead of panicking
- **WHEN** the module ABI encoder initializes deterministic ERC-7579 argument definitions
- **THEN** initialization failures SHALL be represented as returned encoder errors
- **AND** successful install/uninstall calldata encoding SHALL preserve existing selectors and layout
