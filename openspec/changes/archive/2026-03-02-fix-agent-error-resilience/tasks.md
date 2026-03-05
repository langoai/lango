## 1. Gemini Content Sanitization Pipeline

- [x] 1.1 Create internal/provider/gemini/sanitize.go with sanitizeContents function
- [x] 1.2 Implement dropLeadingOrphanedFunctionResponses helper
- [x] 1.3 Implement mergeConsecutiveRoles with shallow clone
- [x] 1.4 Implement ensureFunctionResponsePairs with synthetic FunctionResponse insertion
- [x] 1.5 Add synthetic user turn prepend for model-first sequences
- [x] 1.6 Wire sanitizeContents into gemini.go before GenerateContentStream call
- [x] 1.7 Create sanitize_test.go with 10 table-driven tests and invariant assertions

## 2. Session Event Defense-in-Depth

- [x] 2.1 Add consecutive role merging in EventsAdapter.All() using pending-event buffer pattern
- [x] 2.2 Update Len() to use cached merged events for consistency with All()
- [x] 2.3 Update existing state_test.go tests for alternating roles
- [x] 2.4 Add ConsecutiveRoleMerging tests (merge, no-merge, Len/All consistency)

## 3. Turn Counting Rework

- [x] 3.1 Implement isDelegationEvent helper checking TransferToAgent
- [x] 3.2 Exclude delegation events from turn counting in agent.Run()
- [x] 3.3 Add graceful wrap-up turn logic with wrapUpGranted flag
- [x] 3.4 Add 80% threshold warning logging (single-fire)
- [x] 3.5 Add tests for hasFunctionCalls and isDelegationEvent in agent_test.go

## 4. Multi-Agent Turn Limit Default

- [x] 4.1 Update wiring.go to default maxTurns=50 when multiAgent=true and no explicit config

## 5. Verification

- [x] 5.1 go build ./... passes
- [x] 5.2 go test ./... all tests pass
