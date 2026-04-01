## ADDED Requirements

### Requirement: PhaseTimingEntry type records bootstrap phase duration
The `bootstrap` package SHALL export a `PhaseTimingEntry` struct with `Phase` (string) and `Duration` (time.Duration) fields, both JSON-tagged.

#### Scenario: PhaseTimingEntry struct is usable
- **WHEN** a caller creates a `PhaseTimingEntry{Phase: "openDB", Duration: 50*time.Millisecond}`
- **THEN** the struct holds the phase name and duration and is JSON-serializable

### Requirement: ModuleTimingEntry type records module initialization duration
The `appinit` package SHALL export a `ModuleTimingEntry` struct with `Module` (string) and `Duration` (time.Duration) fields, both JSON-tagged.

#### Scenario: ModuleTimingEntry struct is usable
- **WHEN** a caller creates a `ModuleTimingEntry{Module: "knowledge", Duration: 120*time.Millisecond}`
- **THEN** the struct holds the module name and duration and is JSON-serializable
