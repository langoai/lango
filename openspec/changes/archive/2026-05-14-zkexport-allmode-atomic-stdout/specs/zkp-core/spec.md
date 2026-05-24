## ADDED Requirements
### Requirement: All-mode zkexport stdout stays atomic on failure
`zkexport --all` SHALL keep human-readable stdout progress buffered until the full run succeeds. Failed runs SHALL leave stdout empty while still reporting the failure on stderr and cleaning up files created during the run.

#### Scenario: Later all-mode failure leaves stdout empty
- **WHEN** `zkexport --all` successfully exports at least one earlier circuit and a later circuit export fails
- **THEN** stdout SHALL remain empty for that failed run
- **AND** stderr SHALL report the failing circuit export
- **AND** verifier files created during that run SHALL be removed
