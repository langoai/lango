## ADDED Requirements

### Requirement: Incremental bundle protocol messages
The git bundle protocol SHALL support four new request types: push_incremental_bundle, fetch_incremental, verify_bundle, and has_commit.

#### Scenario: Push incremental bundle
- **WHEN** a push_incremental_bundle request is received with a valid bundle
- **THEN** the handler calls SafeApplyBundle and returns PushBundleResponse with Applied=true

#### Scenario: Fetch incremental bundle
- **WHEN** a fetch_incremental request is received with a base commit hash
- **THEN** the handler calls CreateIncrementalBundle and returns FetchIncrementalResponse with the bundle and HEAD hash

#### Scenario: Verify bundle
- **WHEN** a verify_bundle request is received
- **THEN** the handler calls VerifyBundle and returns VerifyBundleResponse with Valid=true or Valid=false with message

#### Scenario: Has commit check
- **WHEN** a has_commit request is received
- **THEN** the handler calls HasCommit and returns HasCommitResponse with Exists boolean
