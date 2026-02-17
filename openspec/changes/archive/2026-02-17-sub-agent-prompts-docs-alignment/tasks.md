## 1. Per-Agent Default Prompt Files

- [x] 1.1 Create `prompts/agents/operator/IDENTITY.md` matching agentSpecs operator instruction
- [x] 1.2 Create `prompts/agents/navigator/IDENTITY.md` matching agentSpecs navigator instruction
- [x] 1.3 Create `prompts/agents/vault/IDENTITY.md` matching agentSpecs vault instruction
- [x] 1.4 Create `prompts/agents/librarian/IDENTITY.md` matching agentSpecs librarian instruction
- [x] 1.5 Create `prompts/agents/automator/IDENTITY.md` matching agentSpecs automator instruction
- [x] 1.6 Create `prompts/agents/planner/IDENTITY.md` matching agentSpecs planner instruction
- [x] 1.7 Create `prompts/agents/chronicler/IDENTITY.md` matching agentSpecs chronicler instruction

## 2. Global Prompt Updates

- [x] 2.1 Add Cron, Background, Workflow tool categories to `prompts/AGENTS.md`
- [x] 2.2 Add Cron Tool, Background Tool, Workflow Tool sections to `prompts/TOOL_USAGE.md`

## 3. README Updates

- [x] 3.1 Add automator to feature list (line 13)
- [x] 3.2 Add automator to directory structure comment (line 174)
- [x] 3.3 Add automator to per-agent prompt customization section (line 373)
- [x] 3.4 Add automator row to multi-agent orchestration table (line 461-468)
- [x] 3.5 Add automator to workflow supported agents text (line 656)

## 4. Docker Alignment

- [x] 4.1 Add commented prompts volume mount to `docker-compose.yml`

## 5. Verification

- [x] 5.1 Run `go build ./...` to verify no build errors
- [x] 5.2 Run `go test ./...` to verify all tests pass
