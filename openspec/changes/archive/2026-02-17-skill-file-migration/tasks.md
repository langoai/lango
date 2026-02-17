## 1. Skill Package Types & Interfaces

- [x] 1.1 Create `internal/skill/types.go` with independent `SkillEntry` struct (no usage tracking)
- [x] 1.2 Create `internal/skill/store.go` with `SkillStore` interface

## 2. SKILL.md Parser & FileSkillStore

- [x] 2.1 Create `internal/skill/parser.go` with YAML frontmatter + markdown body parsing
- [x] 2.2 Create `internal/skill/file_store.go` implementing `SkillStore` via filesystem
- [x] 2.3 Add `EnsureDefaults(fs.FS)` method for embedded skill deployment
- [x] 2.4 Create `internal/skill/parser_test.go` and `internal/skill/file_store_test.go`

## 3. Registry/Executor/Builder Refactoring

- [x] 3.1 Refactor `internal/skill/registry.go` to use `SkillStore` interface instead of `*knowledge.Store`
- [x] 3.2 Refactor `internal/skill/executor.go` to use `skill.SkillEntry` instead of `knowledge.SkillEntry`
- [x] 3.3 Refactor `internal/skill/builder.go` return types to `skill.SkillEntry`
- [x] 3.4 Add `ListActiveSkills(ctx)` method to Registry
- [x] 3.5 Update `registry_test.go`, `executor_test.go`, `builder_test.go`

## 4. Default CLI Skills

- [x] 4.1 Create `skills/embed.go` with `//go:embed **/SKILL.md` and `DefaultFS()` function
- [x] 4.2 Create 30 SKILL.md files for CLI commands (serve, version, doctor, config-*, security-*, memory-*, agent-*, graph-*, cron-*, workflow-*)

## 5. Config Changes

- [x] 5.1 Add `SkillConfig` struct to `internal/config/types.go`
- [x] 5.2 Remove `AutoApproveSkills` and `MaxSkillsPerDay` from `KnowledgeConfig`
- [x] 5.3 Update `internal/config/loader.go` defaults

## 6. Knowledge Store Cleanup

- [x] 6.1 Remove skill methods from `internal/knowledge/store.go` (SaveSkill, GetSkill, ListActiveSkills, etc.)
- [x] 6.2 Remove `SkillEntry` from `internal/knowledge/types.go`
- [x] 6.3 Remove `maxSkillsPerDay` parameter from `NewStore` signature
- [x] 6.4 Update `internal/knowledge/store_test.go` (remove skill tests, fix NewStore calls)

## 7. ContextRetriever Decoupling

- [x] 7.1 Add `SkillProvider` interface and `SkillInfo` type to `internal/knowledge/retriever.go`
- [x] 7.2 Update `retrieveSkills()` to use `SkillProvider` instead of `store.ListActiveSkills`
- [x] 7.3 Create `skillProviderAdapter` in `internal/app/wiring.go`

## 8. Ent Schema Deletion

- [x] 8.1 Delete `internal/ent/schema/skill.go`
- [x] 8.2 Run `go generate ./internal/ent/...` to regenerate Ent code

## 9. Wiring & App Changes

- [x] 9.1 Create `initSkills()` function in `internal/app/wiring.go`
- [x] 9.2 Separate skills initialization from knowledge in `internal/app/app.go`
- [x] 9.3 Update `buildMetaTools` signature (remove `autoApprove` param) in `internal/app/tools.go`
- [x] 9.4 Wire `SkillProvider` adapter into `initAgent`

## 10. Settings/Onboard Cleanup

- [x] 10.1 Remove `knowledge_auto_approve` and `knowledge_max_skills_day` from `internal/cli/settings/forms_impl.go`
- [x] 10.2 Remove corresponding cases from `internal/cli/tuicore/state_update.go`
- [x] 10.3 Update `internal/cli/settings/forms_impl_test.go`

## 11. Learning Test Fixes

- [x] 11.1 Update `knowledge.NewStore` calls in `internal/learning/` test files (5→4 args)

## 12. Verification

- [x] 12.1 `go build ./...` passes
- [x] 12.2 `go test ./...` passes
