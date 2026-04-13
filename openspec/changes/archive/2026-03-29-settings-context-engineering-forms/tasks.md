# Tasks

## Config defaults
- [x] Add Context allocation defaults to DefaultConfig() in loader.go

## Settings TUI forms
- [x] NewContextProfileForm — ctx_profile select field
- [x] NewRetrievalForm — retrieval_enabled + retrieval_feedback (independent, no VisibleWhen)
- [x] NewAutoAdjustForm — 8 fields with conditional visibility and float validation
- [x] NewContextBudgetForm �� model window, response reserve, 5 allocation ratios (±0.001 validation)
- [x] Register 4 categories in menu.go
- [x] Register 4 cases in setup_flow.go createFormForCategory
- [x] Register 4 cases in editor.go categoryIsEnabled
- [x] Add retrieval → knowledge dependency in dependencies.go
- [x] Add ~22 field handlers in tuicore/state_update.go
- [x] Add 8 form tests in forms_impl_test.go

## Doctor validation
- [x] Create RetrievalCheck with 7 validation rules
- [x] Register RetrievalCheck in AllChecks()
- [x] Enhance ContextHealthCheck: allocation sum ±0.001, RAG without provider
- [x] Add RetrievalCheck tests (7 cases)
- [x] Add ContextHealthCheck enhancement tests

## Downstream
- [x] Update settings.go --help with new categories
- [x] Update doctor.go --help with retrieval/budget checks
- [x] Update README.md config table (Context Profile, Retrieval, Context Budget)
- [x] Update docs/configuration.md (3 new sections with JSON examples)
- [x] Update docs/cli/core.md (settings categories + doctor checks)

## Pre-existing fixes
- [x] Fix configstore/migrate.go Load→Config return type mismatch
