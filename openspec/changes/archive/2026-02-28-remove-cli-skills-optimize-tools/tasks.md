## 1. Remove Default CLI Wrapper Skills

- [x] 1.1 Delete all 42 `skills/<name>/` directories containing lango CLI wrapper SKILL.md files
- [x] 1.2 Create `skills/.placeholder/SKILL.md` without valid YAML frontmatter to satisfy `go:embed **/SKILL.md` pattern
- [x] 1.3 Verify `skills/embed.go` compiles with only the placeholder present

## 2. Optimize Prompts for Tool Priority

- [x] 2.1 Add "Tool Selection Priority" section to `prompts/TOOL_USAGE.md` before the "Exec Tool" section
- [x] 2.2 Add tool selection directive to `prompts/AGENTS.md` before the knowledge system description
- [x] 2.3 Add priority note to "Available Skills" section in `internal/knowledge/retriever.go` AssemblePrompt method

## 3. Verification

- [x] 3.1 Run `go build ./...` and confirm build passes
- [x] 3.2 Run `go test ./...` and confirm all tests pass
- [x] 3.3 Verify `skills/` contains only `embed.go` and `.placeholder/`
