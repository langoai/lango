# Tasks

## loader.go
- [x] Restore LoadResult struct and Load() return type from dev
- [x] Restore collectExplicitKeys, ApplyContextProfile, ResolveContextAutoEnable calls
- [x] Restore Retrieval defaults block
- [x] Restore contextProfile validation in validateConfig
- [x] Add Context allocation defaults (0.30/0.25/0.25/0.10/0.10)

## state_update.go
- [x] Restore 8 Orchestration field handlers from dev
- [x] Restore 9 RunLedger field handlers from dev
- [x] Restore 5 Provenance field handlers from dev
- [x] Restore 9 OS Sandbox field handlers from dev
- [x] Restore 4 TraceStore field handlers from dev
- [x] Re-add 22 Context Engineering field handlers

## setup_flow.go
- [x] Restore runledger, provenance, os_sandbox form factory cases from dev
- [x] Re-add context_profile, retrieval, auto_adjust, context_budget cases

## migrate.go
- [x] Restore LoadResult-based config migration from dev

## Verification
- [x] `go build -tags "fts5,vec" ./...` passes
- [x] `go test -tags "fts5,vec" ./internal/cli/... ./internal/config/... ./internal/configstore/...` all pass
