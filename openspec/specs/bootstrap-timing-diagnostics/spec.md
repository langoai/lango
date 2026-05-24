# Bootstrap Timing Diagnostics

## Purpose
Persist bootstrap phase timing to disk and surface it as a doctor diagnostic check for regression detection.

## Requirements

### Requirement: PhaseTiming file persistence
The bootstrap package SHALL append a JSONL entry to `~/.lango/diagnostics/bootstrap-timing.jsonl` after each successful bootstrap execution. The entry SHALL contain a timestamp (RFC3339), version string, and per-phase timing in milliseconds. The file SHALL be rotated to keep at most N entries (N is a package constant, default 50). Write failures SHALL be logged and SHALL NOT cause bootstrap to fail.

#### Scenario: Successful append
- **WHEN** `Pipeline.Execute` completes successfully
- **THEN** a JSONL line is appended with `ts`, `version`, and `phases` array
- **AND** the diagnostics directory is created if absent

#### Scenario: Rotation at capacity
- **WHEN** the JSONL file contains N entries and a new entry is appended
- **THEN** the oldest entry is removed so the file contains exactly N entries

#### Scenario: Write failure is non-fatal
- **WHEN** the JSONL file cannot be written (permissions, disk full)
- **THEN** the error is logged at warn level
- **AND** bootstrap continues normally

#### Scenario: Corrupted file recovery
- **WHEN** the JSONL file contains malformed lines
- **THEN** the writer discards unreadable lines and continues with valid entries only

### Requirement: BootstrapTimingCheck in doctor
The doctor command SHALL include a `Bootstrap Timing` check implementing the `BootstrapAwareCheck` interface. Current phase timing SHALL come from `boot.PhaseTiming`. Baseline SHALL come from the JSONL file.

#### Scenario: Sufficient baseline — pass
- **WHEN** at least 3 baseline records exist and all current phases are within 2x of baseline median
- **THEN** the check status is `Pass`
- **AND** details show per-phase current vs baseline comparison

#### Scenario: Regression detected — warn
- **WHEN** any current phase duration exceeds 2x the baseline median
- **THEN** the check status is `Warn`
- **AND** the message identifies which phases regressed

#### Scenario: Insufficient baseline — skip
- **WHEN** fewer than 3 baseline records exist in the JSONL file
- **THEN** the check status is `Skip`
- **AND** the message explains more runs are needed for comparison

#### Scenario: Missing or corrupted file — skip
- **WHEN** the JSONL file does not exist or is entirely unreadable
- **THEN** the check status is `Skip`
- **AND** the message explains the file is missing or corrupted

### Requirement: Doctor long description update
The doctor command's long description SHALL include the new `Bootstrap Timing` check in its list and SHALL reflect the updated total check count.

#### Scenario: Help text includes new check
- **WHEN** user runs `lango doctor --help`
- **THEN** the output lists `Bootstrap Timing` under an appropriate category
- **AND** the total check count is incremented by 1
