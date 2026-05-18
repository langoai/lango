# Raise Non-Generated Coverage Baseline

## Summary
Raise Lango's non-generated Go statement coverage to at least 90% while excluding generated code such as Ent output from the target calculation.

## Motivation
The current non-Ent baseline measured on 2026-05-18 is materially below the new production-readiness goal:

- Command: `go test $(go list ./... | grep -v '/internal/ent') -covermode=atomic -coverprofile=/tmp/lango-nonent.cover.out`
- Raw `go tool cover -func` total: 63.9%
- Profile aggregate: 38,563 covered statements / 60,757 total statements / 22,194 uncovered statements

The gap is too large for ad hoc package-by-package fixes. The project needs one umbrella change that defines the coverage measurement contract, excludes generated code intentionally, and raises coverage through focused, reviewable commits.

## Scope
- Define the non-generated coverage measurement contract.
- Add repeatable local coverage reporting that excludes generated Go code.
- Raise test coverage with high-impact package waves prioritized by uncovered statement count.
- Keep commits well-scoped by package or subsystem.
- Use subagent-driven implementation and review checkpoints for each coverage wave.

## Non-Goals
- Do not count Ent-generated code toward the 90% target.
- Do not inflate coverage with tests that only execute code without assertions.
- Do not create separate OpenSpec changes for every package-level test addition.
- Do not refactor production architecture solely to improve coverage percentages.

