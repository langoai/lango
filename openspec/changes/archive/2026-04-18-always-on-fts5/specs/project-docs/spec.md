## ADDED Requirements

### Requirement: Build and installation docs describe FTS5 as always on
Project documentation MUST describe FTS5 as included in the default runtime and MUST NOT require `-tags "fts5"` for normal builds or installs.

#### Scenario: Install docs use default build commands
- **WHEN** a user reads installation or development build instructions
- **THEN** normal build and install examples omit `-tags "fts5"`
- **AND** optional `vec` examples remain explicitly tagged
